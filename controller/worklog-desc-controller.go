package controller

import (
	"fmt"
	"strings"
	"sync"
	"tui/services"
	"tui/utils"

	termhandler "tui/term-handler"
)

type WorklogDescProps struct {
	Width      int
	Height     int
	RenderPosX int
	RenderPosY int
	Title      *string
}

type WorklogDescController struct {
	handler   termhandler.TermhandlerType
	service   services.ServiceType
	mutex     *sync.Mutex
	localChan chan string
	isActive  bool
	wdCursor  int
	offsite   int
	props     WorklogDescProps
	logsData  []services.Logs
}

func NewWorklogDescController(
	handler *termhandler.TermhandlerType,
	service *services.ServiceType,
	mutex *sync.Mutex,
	worklogDescProps WorklogDescProps,
) WorklogDescControllerType {
	return &WorklogDescController{
		handler:   *handler,
		service:   *service,
		mutex:     mutex,
		localChan: make(chan string, 2),
		isActive:  false,
		wdCursor:  0,
		offsite:   0,
		props:     worklogDescProps,
	}
}

// GetChan implements WorklogDescControllerType.
func (w *WorklogDescController) GetChan() chan<- string {
	return w.localChan
}

func (w *WorklogDescController) ReloadData(dateCursor int) {
	w.wdCursor = 0
	w.offsite = 0
	w.cleanBody()

	wkData := w.service.GetWorklogs()
	wlList, ok := wkData.Data[dateCursor]
	if !ok {
		w.logsData = []services.Logs{}
		return
	}

	w.mutex.Lock()
	w.logsData = wlList.Logs
	w.renderBody()
	w.mutex.Unlock()
}

// ListenFromController implements WorklogDescControllerType.
func (w *WorklogDescController) ListenFromController() {
	go func() {
		for resChan := range w.localChan {
			switch resChan {
			case GoUp:
				if w.wdCursor == 0 && w.offsite == 0 {
					continue
				}

				if w.wdCursor == 0 && w.offsite > 0 {
					w.offsite--
					w.renderBody()
					continue
				}

				w.wdCursor--
				w.renderBody()
			case GoDown:
				if w.offsite+w.wdCursor >= len(w.logsData)-1 {
					continue
				}

				if w.wdCursor >= w.props.Height-1 {
					w.offsite++
					w.renderBody()
					continue
				}

				w.wdCursor++
				w.renderBody()
			case QuitDesc:
				w.isActive = false
				w.mutex.Lock()
				w.CreateWindow()
				w.mutex.Unlock()
			case ResetCursor:
				w.isActive = true
				w.wdCursor = 0
				w.offsite = 0
				w.mutex.Lock()
				w.CreateWindow()
				w.mutex.Unlock()
			case Resize:
				w.mutex.Lock()
				w.CreateWindow()
				w.handler.Render()
				w.mutex.Unlock()
			}
		}
	}()
}

// CreateWindow implements WorklogDescControllerType.
func (w *WorklogDescController) CreateWindow() {
	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX,
			w.props.RenderPosY,
		},
	)

	for i := 0; i < w.props.Width; i++ {
		if i == 0 {
			w.handler.Draw("╭")
			continue
		}

		if i == 3 {
			hightlight := "37"
			if w.isActive {
				hightlight = "34"
			}
			printTitle := fmt.Sprintf("\033[%s;1m 󱡠 %s\033[0m ", hightlight, *w.props.Title)
			w.handler.Draw(printTitle)
			i = i + len(*w.props.Title) + 3
			continue
		}

		if i == 20 {
			w.handler.Draw("╥")
			continue
		}

		if i == w.props.Width-1 {
			w.handler.Draw("╮")
			continue
		}

		w.handler.Draw("─")
	}

	w.renderBody()

	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX,
			w.props.RenderPosY + w.props.Height + 1,
		},
	)

	for i := 0; i < w.props.Width; i++ {
		if i == 0 {
			w.handler.Draw("╰")
			continue
		}

		if i == w.props.Width-1 {
			w.handler.Draw("╯")
			continue
		}

		if i == 20 {
			w.handler.Draw("╨")
			continue
		}

		w.handler.Draw("─")
	}
}

// renderBody implements WorklogDescControllerType.
func (w *WorklogDescController) renderBody() {
	highlight := ""
	emptyDesc := ""

	for i := 0; i < w.props.Height; i++ {
		w.handler.MoveCursor(
			termhandler.Position{
				w.props.RenderPosX,
				w.props.RenderPosY + i + 1,
			},
		)

		w.handler.Draw("│")

		if w.wdCursor == i && len(w.logsData) > 0 {
			highlight = "\033[37;44;1m"
		} else {
			highlight = ""
		}

		timeRange := ""
		if len(w.logsData) > 0 && i < len(w.logsData) {
			timeRange = w.logsData[i+w.offsite].TimeRange
		}
		w.handler.Draw(fmt.Sprintf("%s   %s   ", highlight, timeRange))
		w.handler.Draw("\033[0m")

		w.handler.MoveCursor(
			termhandler.Position{
				w.props.RenderPosX + 20,
				w.props.RenderPosY + i + 1,
			},
		)
		w.handler.Draw("║")

		emptyDesc = strings.Repeat(" ", w.props.Width-23)
		w.handler.Draw(emptyDesc)
		if w.wdCursor == i && len(w.logsData) > 0 {
			emptyDesc = w.logsData[i+w.offsite].Comment
			descs := utils.FormatCommentDesc(emptyDesc, 88, '\n')

			for i, desc := range descs {
				w.handler.MoveCursor(
					termhandler.Position{
						w.props.RenderPosX + 22,
						w.props.RenderPosY + i + 1,
					},
				)
				w.handler.Draw(desc)
			}
		}

		w.handler.MoveCursor(
			termhandler.Position{
				w.props.RenderPosX + w.props.Width - 1,
				w.props.RenderPosY + 1 + i,
			},
		)
		w.handler.Draw("│")
		w.handler.Render()
	}
}

func (w *WorklogDescController) cleanBody() {
	for i := 0; i < w.props.Height; i++ {
		w.handler.MoveCursor(
			termhandler.Position{
				w.props.RenderPosX + 1,
				w.props.RenderPosY + i + 1,
			},
		)

		w.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 19)))

		w.handler.MoveCursor(
			termhandler.Position{
				w.props.RenderPosX + 22,
				w.props.RenderPosY + i + 1,
			},
		)
		w.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", w.props.Width-23)))
	}
	w.handler.Render()
}
