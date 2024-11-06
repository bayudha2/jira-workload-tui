package controller

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"tui/services"

	"github.com/mattn/go-tty"

	termhandler "tui/term-handler"
)

const (
	GoUp        string = "up"
	GoDown      string = "down"
	GoLeft      string = "left"
	GoRight     string = "right"
	Resize      string = "resize"
	Search      string = "search"
	QuitDesc    string = "quit_desc"
	ResetCursor string = "reset_cursor"
	LoadingData string = "loading_data"
	ReloadData  string = "reload_data"
)

type ControllerChild map[int]chan<- string

type Controller struct {
	ActiveWidget      int
	controllersChild  ControllerChild
	GlobalChan        chan interface{}
	handler           termhandler.TermhandlerType
	wg                *sync.WaitGroup
	service           services.ServiceType
	channelIsFetching map[int]bool
	getSelectedName   func() string
	getDates          func() (int, int)
}

func NewController(
	wg *sync.WaitGroup,
	handler *termhandler.TermhandlerType,
	service *services.ServiceType,
	ctrlChild ControllerChild,
	globalChan chan interface{},
	getSelectedName func() string,
	getDates func() (int, int),
) ControllerType {
	return &Controller{
		wg:                wg,
		handler:           *handler,
		service:           *service,
		controllersChild:  ctrlChild,
		GlobalChan:        globalChan,
		ActiveWidget:      0,
		channelIsFetching: map[int]bool{},
		getSelectedName:   getSelectedName,
		getDates:          getDates,
	}
}

// ListenExit implements ControllerType.
func (c *Controller) listenExit() {
	csign := make(chan os.Signal, 1)
	signal.Notify(csign, os.Interrupt)

	go func() {
		for range csign {
			c.wg.Done() // <- wg for main go
			c.exitApp()
		}
	}()
}

func (c *Controller) listenRelistenKeyPress() {
	go func() {
		for range c.GlobalChan {
			c.ListenKeyPress()
		}
	}()
}

// ListenResize implements ControllerType.
func (c *Controller) listenResize() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)

	go func() {
		for range sigs {
			c.handler.Clear()
			for _, cchan := range c.controllersChild {
				cchan <- Resize
			}
		}
	}()
}

func (c *Controller) SetupController() {
	c.listenExit()
	c.listenResize()
	c.listenRelistenKeyPress()
	c.ListenKeyPress()
}

// ListenKeyPress implements ControllerType.
func (c *Controller) ListenKeyPress() {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		// BUG: (Goroutine leak): panic when changing widget and move too fast
		// still debuging root cause
		// Possible fix: add recover to panic goroutine function (but which one T_T)
		char, err := tty.ReadRune()
		if err != nil {
			panic(err)
		}

		childChan, ok := c.controllersChild[c.ActiveWidget]
		if !ok {
			continue
		}

		switch char {
		case '1', '2', '3':
			if c.ActiveWidget == 3 {
				continue
			}

			// temporary fix gorutine leaks
			if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
				continue
			}

			charString := string(char)
			charNum, err := strconv.Atoi(charString)
			if err != nil {
				continue
			}

			guideW, _ := c.controllersChild[5]
			guideW <- charString
			c.ActiveWidget = charNum - 1
			for _, cchild := range c.controllersChild {
				go func() {
					cchild <- charString
				}()
			}
		case 'A', 'k':
			if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
				continue
			}
			childChan <- GoUp
		case 'B', 'j':
			if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
				continue
			}
			childChan <- GoDown
		case 'C', 'l':
			if c.ActiveWidget != 1 && c.ActiveWidget != 2 {
				continue
			}

			if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
				continue
			}

			childChan <- GoRight
		case 'D', 'h':
			if c.ActiveWidget != 1 && c.ActiveWidget != 2 {
				continue
			}

			if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
				continue
			}

			childChan <- GoLeft
		case 'r':
			isFetching, ok := c.channelIsFetching[c.ActiveWidget]
			if (c.ActiveWidget != 0 && c.ActiveWidget != 2) || (ok && isFetching) {
				continue
			}

			c.channelIsFetching[c.ActiveWidget] = true

			go func(aw int) {
				childChan <- LoadingData

				switch aw {
				case 0:
					c.service.FetchMembers()
					c.service.FetchUsers()
				case 2:
					month, year := c.getDates()
					name := c.getSelectedName()
					c.service.FetchIssues(services.FetchWorklogPayload{
						Year:  year,
						Month: month,
						Name:  name,
					})
				}

				childChan <- ReloadData
				c.channelIsFetching[aw] = false
			}(c.ActiveWidget)
		case 'q':
			c.exitApp()
		case 13: // handle Enter
			if c.ActiveWidget == 0 || c.ActiveWidget == 1 {
				if isFetching, ok := c.channelIsFetching[c.ActiveWidget]; ok && isFetching {
					continue
				}

				month, year := c.getDates()
				name := c.getSelectedName()
				c.channelIsFetching[c.ActiveWidget] = true

				go func(aw int) {
					wdChan, _ := c.controllersChild[2]
					dashChan, _ := c.controllersChild[4]
					wdChan <- LoadingData
					c.service.FetchIssues(services.FetchWorklogPayload{
						Year:  year,
						Month: month,
						Name:  name,
					})
					wdChan <- ReloadData
					dashChan <- ReloadData
					c.channelIsFetching[aw] = false
				}(c.ActiveWidget)
				continue
			}

			if c.ActiveWidget == 2 {
				c.ActiveWidget = 3
				wdChan, _ := c.controllersChild[3]
				guideW, _ := c.controllersChild[5]

				wdChan <- ResetCursor
				guideW <- "4"

				continue
			}

			if c.ActiveWidget == 3 {
				guideW, _ := c.controllersChild[5]
				childChan <- QuitDesc
				guideW <- "3"
				c.ActiveWidget = 2
			}
		case 'i':
			if c.ActiveWidget == 0 {
				childChan <- Search
				tty.Close()
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// exitApp implements ControllerType.
func (c *Controller) exitApp() {
	c.handler.Clear()
	c.handler.ShowCursor()
	c.handler.MoveCursor(termhandler.Position{1, 1})

	if err := c.handler.Render(); err != nil {
		panic(err)
	}

	os.Exit(0)
}
