package controller

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"tui/services"
	"tui/utils"

	"github.com/mattn/go-tty"

	termhandler "tui/term-handler"
)

type WindowProps struct {
	RenderPosX   int
	RenderPosY   int
	WindowHeight int
	WindowWidth  int
	Title        *string
}

type UserController struct {
	ActiveCursor int
	Offsite      int
	widgetNumber int
	isActive     bool
	localChan    chan string
	globalChan   chan interface{}
	loadingChan  chan struct{}
	Items        []string
	windowProps  WindowProps
	handler      termhandler.TermhandlerType
	service      services.ServiceType
	mutex        *sync.Mutex
}

func NewUserController(
	handler *termhandler.TermhandlerType,
	service *services.ServiceType,
	mutext *sync.Mutex,
	windowProps WindowProps,
	globalChan chan interface{},
	widgetNumeber int,
) WindowControllerType {
	return &UserController{
		ActiveCursor: 0,
		Offsite:      0,
		localChan:    make(chan string, 2),
		globalChan:   globalChan,
		loadingChan:  make(chan struct{}),
		isActive:     true,
		widgetNumber: widgetNumeber,
		windowProps:  windowProps,
		Items:        []string{},
		handler:      *handler,
		service:      *service,
		mutex:        mutext,
	}
}

func (c *UserController) GetActiveCursor() int {
	return c.ActiveCursor
}

func (c *UserController) GetOffsite() int {
	return c.Offsite
}

func (c *UserController) GetChan() chan<- string {
	return c.localChan
}

func (c *UserController) GetSelectedName() string {
	name := ""
	if len(c.Items) > 0 {
		name = c.Items[c.ActiveCursor+c.Offsite]
	}
	return name
}

func (c *UserController) ListenFromController() {
	// getting data from service
	usersName := c.service.GetUsersName()
	c.Items = usersName

	go func() {
		for resChan := range c.localChan {
			switch resChan {
			case "1", "2", "3":
				charNum, err := strconv.Atoi(resChan)
				if err != nil {
					continue
				}

				if c.widgetNumber == (charNum - 1) {
					c.isActive = true
				} else {
					c.isActive = false
				}

				c.mutex.Lock()
				c.reloadActiveIndicator()
				c.mutex.Unlock()
			case GoUp:
				if c.ActiveCursor == 0 && c.Offsite == 0 {
					continue
				}

				if c.ActiveCursor == 0 && c.Offsite > 0 {
					c.mutex.Lock()
					c.Offsite -= 1
					go c.renderList()
					c.mutex.Unlock()
					continue
				}

				c.mutex.Lock()
				c.ActiveCursor -= 1
				go c.renderList()
				c.mutex.Unlock()
			case GoDown:
				if c.Offsite+c.ActiveCursor >= len(c.Items)-1 {
					continue
				}

				if c.ActiveCursor >= c.windowProps.WindowHeight-1 {
					c.mutex.Lock()
					c.Offsite += 1
					go c.renderList()
					c.mutex.Unlock()
					continue
				}

				c.mutex.Lock()
				c.ActiveCursor += 1
				go c.renderList()
				c.mutex.Unlock()
			case LoadingData:
				c.Items = []string{}
				c.Offsite = 0
				c.ActiveCursor = 0

				c.mutex.Lock()
				c.cleanBody()
				c.renderReload()
				c.mutex.Unlock()
			case ReloadData:
				c.loadingChan <- struct{}{}

				usersName := c.service.GetUsersName()
				c.Items = usersName
				c.mutex.Lock()
				go c.renderList()
				c.mutex.Unlock()
			case Search:
				c.handleSearch()
			case Resize:
				c.mutex.Lock()
				c.CreateWindow() // <- current widget
				c.handler.Render()
				c.mutex.Unlock()
			}
		}
	}()
}

func (c *UserController) handleSearch() {
	var input string
	var offsite int

	adjustCursorToSearch := func() {
		c.handler.MoveCursor(
			termhandler.Position{
				c.windowProps.RenderPosX + 9,
				c.windowProps.RenderPosY + 1,
			},
		)
	}

	resetSearch := func() {
		adjustCursorToSearch()
		empytSearch := strings.Repeat(" ", c.windowProps.WindowWidth-10)
		c.handler.Draw(empytSearch)
		c.handler.Render()
		adjustCursorToSearch()
	}

	drawRender := func() {
		c.handler.Draw(input[offsite:])
		c.handler.Render()
	}

	resetSearch()
	c.handler.ShowCursor()
	c.handler.Render()

	tty, err := tty.Open()
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	for {
		adjustCursorToSearch()
		char, err := tty.ReadRune()
		if err != nil {
			panic(err)
		}

		switch char {
		case 9: // Tab
			continue
		case 13: // Enter
			c.handler.HideCursor()
			c.handler.Render()
			c.ActiveCursor = 0
			c.Offsite = 0

			c.Items = utils.FilterStrings(c.service.GetUsersName(), input)
			c.cleanBody()
			c.renderList()

			c.globalChan <- struct{}{}
			return
		case 27: // WARN: Esc and Arrow keys exit search, until maintainer figure out the way
			c.handler.HideCursor()
			resetSearch()
			c.globalChan <- struct{}{}
			return
		case 127:
			if len(input) == 0 {
				continue
			}

			if offsite > 0 {
				offsite -= 1
			}

			tempByteInp := []byte(input)
			tempByteInp[len(input)-1] = byte(' ')
			input = string(tempByteInp)

			drawRender()

			input = input[:len(input)-1]
			adjustCursorToSearch()
		default:
			input += string(char)
			if len(input) > c.windowProps.WindowWidth-10 {
				offsite += 1
			}
		}

		drawRender()
	}
}

func (c *UserController) CreateWindow() {
	// render top border
	c.handler.MoveCursor(termhandler.Position{c.windowProps.RenderPosX, c.windowProps.RenderPosY})
	for i := 0; i < c.windowProps.WindowWidth-2; i++ {
		if i == 0 {
			c.handler.Draw("╭")
			continue
		}

		if i == 3 {
			printTitle := fmt.Sprintf(" \u001b[37;1m %s\033[0m ", *c.windowProps.Title)
			c.handler.Draw(printTitle)
			i = i + len(*c.windowProps.Title) + 1
			continue
		}

		c.handler.Draw("─")

		if i == c.windowProps.WindowWidth-3 {
			c.handler.Draw("╮")
		}
	}

	// render body
	c.renderList()

	// render bottom border
	c.handler.MoveCursor(
		termhandler.Position{
			c.windowProps.RenderPosX,
			c.windowProps.RenderPosY + c.windowProps.WindowHeight + 3,
		},
	)
	for i := 0; i < c.windowProps.WindowWidth; i++ {
		if i == 0 {
			c.handler.Draw("╰")
			continue
		}

		if i == c.windowProps.WindowWidth-4 {
			hightlight := "\u001b[32;1m"
			c.handler.Draw(fmt.Sprintf(" %s%d\033[0m ", hightlight, c.widgetNumber+1))
			i = c.windowProps.WindowWidth - 2
		} else {
			c.handler.Draw("─")
		}

		if i == c.windowProps.WindowWidth-1 {
			c.handler.Draw("╯")
		}
	}
}

func (c *UserController) renderSearch() {
	c.handler.MoveCursor(
		termhandler.Position{c.windowProps.RenderPosX, c.windowProps.RenderPosY + 1},
	)
	c.handler.Draw("│")
	c.handler.Draw("Search:")
	c.handler.MoveCursor(
		termhandler.Position{
			c.windowProps.RenderPosX + c.windowProps.WindowWidth,
			c.windowProps.RenderPosY + 1,
		},
	)
	c.handler.Draw("│")

	c.handler.MoveCursor(
		termhandler.Position{c.windowProps.RenderPosX, c.windowProps.RenderPosY + 2},
	)

	for i := 0; i < c.windowProps.WindowWidth; i++ {
		if i == 0 {
			c.handler.Draw("├")
			continue
		}

		c.handler.Draw("─")

		if i == c.windowProps.WindowWidth-1 {
			c.handler.Draw("┤")
		}
	}
}

func (c *UserController) renderList() {
	hightlight := "\u001b[30;107m"

	c.renderSearch() // create search input

	for i := 0; i < c.windowProps.WindowHeight; i++ {
		c.handler.MoveCursor(
			termhandler.Position{
				c.windowProps.RenderPosX,
				c.windowProps.RenderPosY + i + 3,
			},
		)
		c.handler.Draw("│")

		if len(c.Items) > 0 {
			item := ""
			c.renderBodyList(&item, i, hightlight)
			c.handler.Draw(item)
			c.handler.Draw("\033[0m")
		}

		c.handler.MoveCursor(
			termhandler.Position{
				c.windowProps.RenderPosX + c.windowProps.WindowWidth,
				c.windowProps.RenderPosY + i + 3,
			},
		)
		c.handler.Draw("│")
	}

	if err := c.handler.Render(); err != nil {
		panic(err)
	}
}

func (c *UserController) renderBodyList(item *string, i int, hightlight string) {
	if i > len(c.Items)-1 {
		return
	}

	*item = fmt.Sprintf("%s", c.Items[i+c.Offsite])
	if len(*item) > c.windowProps.WindowWidth {
		temStr := *item
		*item = temStr[:c.windowProps.WindowWidth-1]
	}

	if i == c.ActiveCursor {
		*item = hightlight + " " + *item
	}

	remSpace := c.windowProps.WindowWidth - len(c.Items[i+c.Offsite]) - 1 // 2: icon + space
	filler := ""

	if remSpace > 0 {
		filler = strings.Repeat(" ", remSpace)
	}
	*item = *item + filler
}

func (c *UserController) cleanBody() {
	for i := 0; i < c.windowProps.WindowHeight; i++ {
		c.handler.MoveCursor(
			termhandler.Position{
				c.windowProps.RenderPosX + 1,
				c.windowProps.RenderPosY + i + 3,
			},
		)

		emptySpace := strings.Repeat(" ", c.windowProps.WindowWidth-1)
		c.handler.Draw(emptySpace)
	}

	c.handler.Render()
}

func (c *UserController) reloadActiveIndicator() {
	c.handler.MoveCursor(
		termhandler.Position{
			c.windowProps.RenderPosX + c.windowProps.WindowWidth - 3,
			c.windowProps.RenderPosY + c.windowProps.WindowHeight + 3,
		},
	)

	highlight := ""
	if c.isActive {
		highlight = "\u001b[32;1m"
	}
	c.handler.Draw(fmt.Sprintf("%s1\033[0m", highlight))
	c.handler.Render()
}

func (c *UserController) renderReload() {
	go func() {
		loading := []string{"|", "/", "-", "\\"}
		loadingIndex := 0

		for {
			select {
			case <-c.loadingChan:
				return
			case <-time.After(100 * time.Millisecond):
				c.handler.MoveCursor(
					termhandler.Position{
						c.windowProps.RenderPosX + 1,
						c.windowProps.RenderPosY + 3,
					},
				)

				c.handler.Draw(fmt.Sprintf("Loading... %s", loading[loadingIndex]))
				c.handler.Render()

				if loadingIndex == len(loading)-1 {
					loadingIndex = 0
				} else {
					loadingIndex++
				}
			}
		}
	}()
}
