package controller

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"tui/services"
	"tui/utils"

	termhandler "tui/term-handler"
)

type WorklogProps struct {
	Width      int
	Height     int
	RenderPosX int
	RenderPosY int
	Title      *string
}

type DateCursorTrack struct {
	current [2]int
	before  [2]int
}

type WorklogData struct {
	date string
	day  string
	data services.FormattedWorklogData
}

type WorklogController struct {
	handler          termhandler.TermhandlerType
	service          services.ServiceType
	mutex            *sync.Mutex
	globalChan       interface{}
	localChan        chan string
	loadingChan      chan struct{}
	widgetNumber     int
	isActive         bool
	isLoading        bool
	dateCursor       int
	dateCursorBefore int
	dateCursorTrack  DateCursorTrack
	props            WorklogProps
	worklogData      []WorklogData
	ReloadWLDesc     func(int)
}

func NewWorklogController(
	handler *termhandler.TermhandlerType,
	service *services.ServiceType,
	mutex *sync.Mutex,
	globalChan interface{},
	widgetNumber int,
	worklogsProps WorklogProps,
	reloadWKDesc func(int),
) WorklogControllerType {
	return &WorklogController{
		handler:          *handler,
		service:          *service,
		mutex:            mutex,
		globalChan:       globalChan,
		localChan:        make(chan string, 2),
		loadingChan:      make(chan struct{}),
		widgetNumber:     widgetNumber,
		isActive:         false,
		isLoading:        false,
		dateCursor:       0,
		dateCursorBefore: 0,
		dateCursorTrack: DateCursorTrack{
			current: [2]int{1, 2},
			before:  [2]int{1, 2},
		},
		props:        worklogsProps,
		ReloadWLDesc: reloadWKDesc,
	}
}

// GetChan implements WorklogControllerType.
func (w *WorklogController) GetChan() chan<- string {
	return w.localChan
}

func (w *WorklogController) GetDateCursor() int {
	return w.dateCursor
}

// ListenFromController implements WorklogControllerType.
func (w *WorklogController) ListenFromController() {
	go func() {
		for resChan := range w.localChan {
			switch resChan {
			case GoUp:
				if w.dateCursor < 7 || w.isLoading {
					continue
				}

				w.mutex.Lock()
				w.dateCursorTrack.before = w.dateCursorTrack.current
				w.dateCursorTrack.current = [2]int{
					w.dateCursorTrack.current[0],
					w.dateCursorTrack.current[1] - 6,
				}

				w.dateCursorBefore = w.dateCursor
				w.dateCursor -= 7
				w.reloadActiveDateIndicator()
				w.mutex.Unlock()

				w.ReloadWLDesc(w.dateCursor)
			case GoDown:
				if (w.dateCursor+7 > len(w.worklogData)-1) || w.isLoading {
					continue
				}

				w.mutex.Lock()
				w.dateCursorTrack.before = w.dateCursorTrack.current
				w.dateCursorTrack.current = [2]int{
					w.dateCursorTrack.current[0],
					w.dateCursorTrack.current[1] + 6,
				}

				w.dateCursorBefore = w.dateCursor
				w.dateCursor += 7
				w.reloadActiveDateIndicator()
				w.mutex.Unlock()

				w.ReloadWLDesc(w.dateCursor)
			case GoLeft:
				if w.dateCursor == 0 || w.isLoading {
					continue
				}

				w.mutex.Lock()
				w.dateCursorTrack.before = w.dateCursorTrack.current
				if w.dateCursor%7 == 0 {
					w.dateCursorTrack.current = [2]int{
						91,
						w.dateCursorTrack.current[1] - 6,
					}
				} else {
					w.dateCursorTrack.current = [2]int{
						w.dateCursorTrack.current[0] - 15,
						w.dateCursorTrack.current[1],
					}
				}

				w.dateCursorBefore = w.dateCursor
				w.dateCursor--
				w.reloadActiveDateIndicator()
				w.mutex.Unlock()

				w.ReloadWLDesc(w.dateCursor)
			case GoRight:
				if (w.dateCursor == len(w.worklogData)-1) || w.isLoading {
					continue
				}

				w.mutex.Lock()
				w.dateCursorTrack.before = w.dateCursorTrack.current
				if (w.dateCursor+1)%7 == 0 {
					w.dateCursorTrack.current = [2]int{
						1,
						w.dateCursorTrack.current[1] + 6,
					}
				} else {
					w.dateCursorTrack.current = [2]int{
						w.dateCursorTrack.current[0] + 15,
						w.dateCursorTrack.current[1],
					}
				}

				w.dateCursorBefore = w.dateCursor
				w.dateCursor++
				w.reloadActiveDateIndicator()
				w.mutex.Unlock()

				w.ReloadWLDesc(w.dateCursor)
			case LoadingData:
				w.dateCursor = 0
				w.dateCursorTrack = DateCursorTrack{
					current: [2]int{1, 2},
					before:  [2]int{1, 2},
				}
				w.worklogData = []WorklogData{}
				w.isLoading = true
				w.renderReload()
			case ReloadData:
				w.loadingChan <- struct{}{}
				w.isLoading = false

				w.mutex.Lock()
				w.ReloadWLDesc(w.dateCursor)
				w.mapWorklogData()
				w.renderBody()
				w.mutex.Unlock()
			case "1", "2", "3":
				charNum, err := strconv.Atoi(resChan)
				if err != nil {
					continue
				}

				if w.widgetNumber == (charNum - 1) {
					w.isActive = true
				} else {
					w.isActive = false
				}

				w.mutex.Lock()
				w.reloadActiveIndicator()
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

// createWindow implements WorklogControllerType.
func (w *WorklogController) CreateWindow() {
	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX,
			w.props.RenderPosY,
		},
	)

	for i := 0; i < w.props.Width; i++ {
		if i == 3 {
			printTitle := fmt.Sprintf("\033[37;1m 󰃰 %s\033[0m ", *w.props.Title)
			w.handler.Draw(printTitle)
			i = i + len(*w.props.Title) + 4
			continue
		}

		if i == w.props.Width-5 {
			printTitle := fmt.Sprintf(" %d ", w.widgetNumber+1)
			w.handler.Draw(printTitle)
			i = i + 1
			continue
		}

		w.handler.Draw("─")
	}

	w.renderBody()
}

func (w *WorklogController) mapWorklogData() {
	w.worklogData = []WorklogData{} // reset data
	wlData := w.service.GetWorklogs()

	for i := 0; i < wlData.LastDate+1; i++ {
		date := fmt.Sprintf("%d", i)
		month := fmt.Sprintf("%d", wlData.Month)

		if i < 10 {
			date = fmt.Sprintf("0%d", i)
		}

		if wlData.Month < 10 {
			month = fmt.Sprintf("0%d", wlData.Month)
		}

		fDate := fmt.Sprintf("%d-%s-%s", wlData.Year, month, date)
		parsed, _ := time.Parse(time.DateOnly, fDate)
		w.worklogData = append(w.worklogData, WorklogData{
			date: date,
			day:  parsed.Weekday().String(),
			data: wlData.Data[i],
		})
	}
}

func (w *WorklogController) renderBody() {
	topLeftCorner := ""
	topRightCorner := ""
	bottomLeftCorner := ""
	dateIndex := 0

	for k := 0; k < 5; k++ {
		for i := 0; i < 7; i++ {
			w.handler.MoveCursor(
				termhandler.Position{
					w.props.RenderPosX + (i * 15),
					w.props.RenderPosY + 1 + (k * 6),
				},
			)

			if k == 0 && i == 0 {
				topLeftCorner = "┌"
			} else if i > 0 && k > 0 {
				topLeftCorner = "┼"
			} else if i == 0 {
				topLeftCorner = "├"
			} else if i > 0 {
				topLeftCorner = "┬"
			}

			if k == 0 && i == 6 {
				topRightCorner = "┐"
			} else {
				topRightCorner = "┤"
			}
			w.handler.Draw(fmt.Sprintf("%s──────────────%s", topLeftCorner, topRightCorner))

			for j := 0; j < 5; j++ {
				w.handler.MoveCursor(
					termhandler.Position{
						w.props.RenderPosX + (i * 15),
						w.props.RenderPosY + 2 + j + (k * 6),
					},
				)

				if j == 0 {
					date := "  "
					day := "  "
					highlight := ""

					if dateIndex < len(w.worklogData) {
						if w.dateCursor == dateIndex {
							highlight = "\033[37;44;1;3m"
						}

						date = fmt.Sprintf("%s%s\033[0m", highlight, w.worklogData[dateIndex].date)
						if !(i == 0 && k == 0) {
							day = fmt.Sprintf("%s", w.worklogData[dateIndex].day)
						}
					}

					remSpace := 12 - len(day)
					filler := ""

					if remSpace > 0 {
						filler = strings.Repeat(" ", remSpace)
					}

					w.handler.Draw(fmt.Sprintf("│%s%s%s│", date, filler, day))
					continue
				}

				if !(i == 0 && k == 0) && j == 4 {
					todayTime := ""
					todayTimeSpent := "  "
					if dateIndex < len(w.worklogData) {
						timeSpent := utils.FormatSecondToHourMinute(
							w.worklogData[dateIndex].data.TimeSpent,
							false,
						)
						tsHighlight := w.calculateTimespentHighlight(
							w.worklogData[dateIndex].data.TimeSpent,
						)
						todayTimeSpent = fmt.Sprintf("%s / 8h", timeSpent)
						todayTime = fmt.Sprintf("\033[%s;1m%s\033[0m/8h", tsHighlight, timeSpent)
					}

					remSpace := 16 - len(todayTimeSpent)
					filler := ""
					if remSpace > 0 {
						filler = strings.Repeat(" ", remSpace)
					}
					w.handler.Draw(fmt.Sprintf("│%s%s│", todayTime, filler))
					continue
				}

				w.handler.Draw(fmt.Sprintf("│%s│", strings.Repeat(" ", 14)))
			}

			w.handler.MoveCursor(
				termhandler.Position{
					w.props.RenderPosX + (i * 15),
					w.props.RenderPosY + 7 + (k * 6),
				},
			)

			if k == 4 && i == 0 {
				bottomLeftCorner = "└"
			} else {
				bottomLeftCorner = "┴"
			}

			w.handler.Draw(fmt.Sprintf("%s──────────────┘", bottomLeftCorner))
			w.handler.Render()

			dateIndex++
		}
	}
}

func (w *WorklogController) calculateTimespentHighlight(n int) string {
	if n > 28800 {
		return "36"
	} else if n > 21600 {
		return "32"
	} else if n > 14400 {
		return "33"
	}
	return "31"
}

func (w *WorklogController) renderReload() {
	go func() {
		loading := []string{"|", "/", "-", "\\"}
		loadingIndex := 0

		for {
			select {
			case <-w.loadingChan:
				return
			case <-time.After(100 * time.Millisecond):
				w.handler.MoveCursor(
					termhandler.Position{
						w.props.RenderPosX + 1,
						w.props.RenderPosY + 2,
					},
				)

				w.handler.Draw(fmt.Sprintf("Loading... %s", loading[loadingIndex]))
				w.handler.Render()

				if loadingIndex == len(loading)-1 {
					loadingIndex = 0
				} else {
					loadingIndex++
				}
			}
		}
	}()
}

func (w *WorklogController) reloadActiveDateIndicator() {
	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX + w.dateCursorTrack.before[0],
			w.props.RenderPosY + w.dateCursorTrack.before[1],
		},
	)

	dateBeforeNum := fmt.Sprintf("%d", w.dateCursorBefore)
	if w.dateCursorBefore < 10 {
		dateBeforeNum = fmt.Sprintf("0%d", w.dateCursorBefore)
	}

	w.handler.Draw(dateBeforeNum)

	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX + w.dateCursorTrack.current[0],
			w.props.RenderPosY + w.dateCursorTrack.current[1],
		},
	)

	highlight := "\033[37;44;1;3m"

	dateNum := fmt.Sprintf("%d", w.dateCursor)
	if w.dateCursor < 10 {
		dateNum = fmt.Sprintf("0%d", w.dateCursor)
	}

	w.handler.Draw(fmt.Sprintf("%s%s\033[0m", highlight, dateNum))

	w.handler.Render()
}

func (w *WorklogController) reloadActiveIndicator() {
	w.handler.MoveCursor(
		termhandler.Position{
			w.props.RenderPosX + w.props.Width - 5,
			w.props.RenderPosY,
		},
	)

	highlight := ""
	if w.isActive {
		highlight = "\u001b[32;1m"
	}
	w.handler.Draw(fmt.Sprintf("%s3\033[0m", highlight))
	w.handler.Render()
}
