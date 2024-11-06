package controller

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	termhandler "tui/term-handler"
)

var Months = [12]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

type DateProps struct {
	RenderPosX int
	RenderPosY int
	Width      int
	Height     int
	Title      *string
}

type DateController struct {
	monthCursor  int
	dateCursor   int
	year         int
	widgetNumber int
	isActive     bool
	localChan    chan string
	globalChan   chan interface{}
	handler      termhandler.TermhandlerType
	mutex        *sync.Mutex
	props        DateProps
}

func NewDateController(
	handler *termhandler.TermhandlerType,
	mutex *sync.Mutex,
	globalChan chan interface{},
	widgetNumber int,
	dateProps DateProps,
) DateControllerType {
	return &DateController{
		monthCursor:  int(time.Now().Month()) - 1,
		dateCursor:   0,
		year:         time.Now().Year(),
		localChan:    make(chan string, 2),
		globalChan:   globalChan,
		widgetNumber: widgetNumber,
		isActive:     false,
		handler:      *handler,
		mutex:        mutex,
		props:        dateProps,
	}
}

func (d *DateController) GetChan() chan<- string {
	return d.localChan
}

func (d *DateController) GetDates() (int, int) {
	return d.monthCursor + 1, d.year
}

func (d *DateController) ListenFromController() {
	go func() {
		for resChan := range d.localChan {
			switch resChan {
			case "1", "2", "3":
				charNum, err := strconv.Atoi(resChan)
				if err != nil {
					continue
				}

				if d.widgetNumber == (charNum - 1) {
					d.isActive = true
				} else {
					d.isActive = false
				}

				d.mutex.Lock()
				d.reloadActiveIndicator()
				d.mutex.Unlock()
			case GoUp:
				switch d.dateCursor {
				case 0:
					if d.monthCursor == 0 {
						continue
					}
					d.monthCursor--
				case 1:
					d.year--
				}

				d.mutex.Lock()
				d.renderBody()
				d.mutex.Unlock()
			case GoLeft:
				if d.dateCursor == 0 {
					continue
				}

				d.mutex.Lock()
				d.dateCursor--
				d.renderBody()
				d.mutex.Unlock()
			case GoRight:
				if d.dateCursor == 1 {
					continue
				}

				d.mutex.Lock()
				d.dateCursor++
				d.renderBody()
				d.mutex.Unlock()
			case GoDown:
				switch d.dateCursor {
				case 0:
					if d.monthCursor == 11 {
						continue
					}
					d.monthCursor++
				case 1:
					d.year++
				}

				d.mutex.Lock()
				d.renderBody()
				d.mutex.Unlock()
			case Resize:
				d.mutex.Lock()
				d.CreateWindow()
				d.handler.Render()
				d.mutex.Unlock()
			}
		}
	}()
}

func (d *DateController) CreateWindow() {
	// render top border
	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX, d.props.RenderPosY})
	for i := 0; i < d.props.Width-2; i++ {
		if i == 0 {
			d.handler.Draw("╭")
			continue
		}

		if i == 3 {
			printTitle := fmt.Sprintf(" \u001b[37;1m󰃭 %s\033[0m ", *d.props.Title)
			d.handler.Draw(printTitle)
			i = i + len(*d.props.Title) + 1
			continue
		}

		if i == d.props.Width/2-1 {
			d.handler.Draw("┬") // separator
		} else {
			d.handler.Draw("─")
		}

		if i == d.props.Width-3 {
			d.handler.Draw("╮")
		}
	}

	// render body
	d.renderBody()

	// render bottom border
	d.handler.MoveCursor(
		termhandler.Position{
			d.props.RenderPosX,
			d.props.RenderPosY + d.props.Height + 1,
		},
	)

	for i := 0; i < d.props.Width; i++ {
		if i == 0 {
			d.handler.Draw("╰")
			continue
		}

		if i == d.props.Width/2+1 {
			d.handler.Draw("┴") // separator
		} else if i == d.props.Width-4 {
			d.handler.Draw(fmt.Sprintf(" %d ", d.widgetNumber+1))
			i = d.props.Width - 2
		} else {
			d.handler.Draw("─")
		}

		if i == d.props.Width-1 {
			d.handler.Draw("╯")
		}
	}
}

func (d *DateController) renderBody() {
	for i := 0; i < d.props.Height; i++ {
		d.handler.MoveCursor(termhandler.Position{
			d.props.RenderPosX,
			d.props.RenderPosY + i + 1,
		})

		d.handler.Draw("│")

		highlight := ""
		if d.isActive && d.dateCursor == 0 {
			highlight = "\u001b[37;1m"
		}

		d.handler.Draw(fmt.Sprintf(" %s%s\033[0m", highlight, Months[d.monthCursor]))
		d.handler.Draw(strings.Repeat(" ", d.props.Width/2-(len(Months[d.monthCursor]))-1))

		d.handler.Draw("│") // separator

		highlight = ""
		if d.isActive && d.dateCursor == 1 {
			highlight = "\u001b[37;1m"
		}

		d.handler.Draw(fmt.Sprintf(" %s%d\033[0m", highlight, d.year))
		d.handler.Draw(strings.Repeat(" ", d.props.Width/2-6))

		d.handler.Draw("│")

		if err := d.handler.Render(); err != nil {
			panic(err)
		}
	}
}

func (d *DateController) reloadActiveIndicator() {
	d.handler.MoveCursor(
		termhandler.Position{
			d.props.RenderPosX + d.props.Width - 3,
			d.props.RenderPosY + d.props.Height + 1,
		},
	)

	highlight := ""
	if d.isActive {
		highlight = "\u001b[32;1m"
	}
	d.handler.Draw(fmt.Sprintf("%s2\033[0m", highlight))
	d.handler.Render()
}
