package controller

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"tui/services"
	"tui/utils"

	termhandler "tui/term-handler"
)

type DashboardProps struct {
	Width      int
	Height     int
	RenderPosX int
	RenderPosY int
	Title      *string
}

type DashboardSummary struct {
	TotalBacklog   int
	TotalWorklog   int
	TotalTimeSpent int
	Name           string
	Email          string
	Month          int
	Year           int
}

type DashboardController struct {
	handler     termhandler.TermhandlerType
	service     services.ServiceType
	mutex       *sync.Mutex
	localChan   chan string
	globalChan  interface{}
	props       DashboardProps
	summaryData DashboardSummary
}

func NewDashboardController(
	handler *termhandler.TermhandlerType,
	service *services.ServiceType,
	mutex *sync.Mutex,
	globalChan chan interface{},
	dashboardProps DashboardProps,
) DashboardControllerType {
	return &DashboardController{
		handler:    *handler,
		service:    *service,
		mutex:      mutex,
		localChan:  make(chan string, 2),
		globalChan: globalChan,
		props:      dashboardProps,
	}
}

// GetChan implements DashboardControllerType.
func (d *DashboardController) GetChan() chan<- string {
	return d.localChan
}

// ListenFromController implements DashboardControllerType.
func (d *DashboardController) ListenFromController() {
	go func() {
		for resChan := range d.localChan {
			switch resChan {
			case ReloadData:
				sl := d.service.GetSummaryLog()
				wl := d.service.GetWorklogs()
				user := d.service.GetUser()
				d.summaryData = DashboardSummary{
					TotalBacklog:   sl.TotalBacklog,
					TotalWorklog:   sl.TotalWorklog,
					TotalTimeSpent: sl.TotalTimeSpent,
					Name:           user.DisplayName,
					Email:          user.EmailAdrres,
					Month:          wl.Month,
					Year:           wl.Year,
				}

				d.mutex.Lock()
				d.cleanBody()
				d.renderBody()
				d.handler.Render()
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

// createWindow implements DashboardControllerType.
func (d *DashboardController) CreateWindow() {
	d.handler.MoveCursor(
		termhandler.Position{
			d.props.RenderPosX,
			d.props.RenderPosY,
		},
	)

	// border top
	for i := 0; i < d.props.Width; i++ {
		if i == 0 {
			d.handler.Draw("╭")
			continue
		}

		if i == d.props.Width-1 {
			d.handler.Draw("╮")
			continue
		}

		if i == 3 {
			printTitle := fmt.Sprintf(" \033[37;1m󰨇 %s\033[0m ", *d.props.Title)
			d.handler.Draw(printTitle)
			i = i + len(*d.props.Title) + 3
			continue
		}

		d.handler.Draw("─")
	}

	d.renderBody()

	// border left
	for i := 0; i < d.props.Height-1; i++ {
		d.handler.MoveCursor(
			termhandler.Position{
				d.props.RenderPosX,
				d.props.RenderPosY + i + 1,
			},
		)
		d.handler.Draw("│")
	}

	// border right
	for i := 0; i < d.props.Height-1; i++ {
		d.handler.MoveCursor(
			termhandler.Position{
				d.props.RenderPosX + d.props.Width - 1,
				d.props.RenderPosY + i + 1,
			},
		)
		d.handler.Draw("│")
	}

	// border bottom
	d.handler.MoveCursor(
		termhandler.Position{
			d.props.RenderPosX,
			d.props.RenderPosY + d.props.Height,
		},
	)
	for i := 0; i < d.props.Width; i++ {
		if i == 0 {
			d.handler.Draw("╰")
			continue
		}

		if i == d.props.Width-1 {
			d.handler.Draw("╯")
			continue
		}

		d.handler.Draw("─")
	}
}

func (d *DashboardController) renderBody() {
	if d.summaryData.Name == "" {
		return
	}
	// title body
	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 1})
	d.handler.Draw(fmt.Sprintf("\033[97;1m%s\033[0m", d.summaryData.Name))

	d.handler.MoveCursor(
		termhandler.Position{d.props.RenderPosX + d.props.Width - 11, d.props.RenderPosY + 1},
	)
	date := time.Date(
		d.summaryData.Year,
		time.Month(d.summaryData.Month+1),
		0,
		0,
		0,
		0,
		0,
		&time.Location{},
	)
	month := ""
	year := ""
	if d.summaryData.Year > 0 {
		year = fmt.Sprintf("%d", date.Year())
		month = date.Format("Jan")
	}

	d.handler.Draw(fmt.Sprintf("\033[37;1m%s %s\033[0m", month, year))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 2})
	email := "-"
	if d.summaryData.Email != "" {
		email = d.summaryData.Email
	}

	d.handler.Draw(fmt.Sprintf("%s", email))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 3})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat("─", d.props.Width-6)))

	// body
	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 4})
	d.handler.Draw(
		fmt.Sprintf("󰤓 Total Backlog       : \033[97;1m%d\033[0m", d.summaryData.TotalBacklog),
	)

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 5})
	d.handler.Draw(
		fmt.Sprintf(" Total Worklog       : \033[97;1m%d\033[0m", d.summaryData.TotalWorklog),
	)

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 6})
	d.handler.Draw(
		fmt.Sprintf(
			"󱑒 Total Time Spent    : \033[97;1m%s\033[0m",
			utils.FormatSecondToHourMinute(d.summaryData.TotalTimeSpent, true),
		),
	)

	// As of Today Percentage
	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 8})
	d.handler.Draw(" \033[97;1mAs of Today\033[0m")

	targetMonth, targetToday := utils.GetWorkDays(d.summaryData.Month, d.summaryData.Year)

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 9})
	d.handler.Draw(fmt.Sprintf(" Target: %s", utils.FormatSecondToHourMinute(targetToday, true)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 10})
	d.handler.Draw(" 0")

	d.handler.MoveCursor(
		termhandler.Position{d.props.RenderPosX + 1 + d.props.Width - 6, d.props.RenderPosY + 10},
	)
	d.handler.Draw("100")

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 11})
	barFiller := d.generateRGBChart(targetToday)
	d.handler.Draw(
		fmt.Sprintf("┃%s\033[0m", barFiller),
	)

	d.handler.MoveCursor(
		termhandler.Position{d.props.RenderPosX + d.props.Width - 3, d.props.RenderPosY + 11},
	)
	d.handler.Draw("┃")

	// Percentage Month

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 13})
	d.handler.Draw(" \033[97;1mMonth\033[0m")

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 14})
	d.handler.Draw(fmt.Sprintf(" Target: %s", utils.FormatSecondToHourMinute(targetMonth, true)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 15})
	d.handler.Draw(" 0")

	d.handler.MoveCursor(
		termhandler.Position{d.props.RenderPosX + 1 + d.props.Width - 6, d.props.RenderPosY + 15},
	)
	d.handler.Draw("100")

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 16})
	barFiller = d.generateRGBChart(targetMonth)
	d.handler.Draw(fmt.Sprintf("┃%s\033[0m", barFiller))

	d.handler.MoveCursor(
		termhandler.Position{d.props.RenderPosX + d.props.Width - 3, d.props.RenderPosY + 16},
	)
	d.handler.Draw("┃")
}

func (d *DashboardController) generateRGBChart(target int) string {
	var barPercentage float32
	fromRGB := [3]int{255, 0, 64}
	toRGB := [3]int{0, 255, 38}
	if target > 0 {
		barPercentage = (float32(d.summaryData.TotalTimeSpent) / float32(target))

		if barPercentage > 1 {
			fromRGB = [3]int{85, 255, 255}
			toRGB = [3]int{85, 255, 255}
		}
	}

	barFiller := ""
	utils.CalculateRGB(&barFiller, '█', d.props.Width-5, fromRGB, toRGB)

	if barPercentage <= 1 {
		i := 0
		endIndex := int(barPercentage * float32(len(barFiller)))

		for {
			if endIndex+i > len(barFiller)-1 || endIndex+(i*-1) < 0 {
				break
			}

			if barFiller[endIndex+i] == 'm' {
				break
			}

			if barFiller[endIndex+(i*-1)] == 'm' {
				i = i * -1
				break
			}
			i++
		}

		endIndex = endIndex + i
		barFiller = barFiller[0:endIndex]
	}

	return barFiller
}

func (d *DashboardController) cleanBody() {
	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 1})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", d.props.Width-10)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 3, d.props.RenderPosY + 2})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", d.props.Width-10)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 27, d.props.RenderPosY + 4})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 20)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 27, d.props.RenderPosY + 5})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 20)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 27, d.props.RenderPosY + 6})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 20)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 10, d.props.RenderPosY + 9})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 20)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 11})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", d.props.Width-4)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 10, d.props.RenderPosY + 14})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 20)))

	d.handler.MoveCursor(termhandler.Position{d.props.RenderPosX + 1, d.props.RenderPosY + 16})
	d.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", d.props.Width-4)))

	d.handler.Render()
}
