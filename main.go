package main

import (
	"sync"
	"tui/config"
	"tui/controller"
	"tui/services"
	"tui/utils"

	termhandler "tui/term-handler"
)

func main() {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// setup config
	cfg := config.NewConfig()

	// setup program
	thandler := termhandler.NewTermHandler()
	thandler.Clear()
	thandler.HideCursor()

	globalChan := make(chan interface{})

	service := services.NewService(&wg, &mutex, &thandler, &cfg)
	service.InitService()

	// user filter widget
	userCtrlr := controller.NewUserController(
		&thandler,
		&service,
		&mutex,
		controller.WindowProps{
			RenderPosX:   2,
			RenderPosY:   8,
			WindowHeight: 12,
			WindowWidth:  31,
			Title:        utils.StrToPtr("Users"),
		},
		globalChan,
		0,
	)

	// dates filter widget
	dateCtrlr := controller.NewDateController(
		&thandler,
		&mutex,
		globalChan,
		1,
		controller.DateProps{
			RenderPosX: 2,
			RenderPosY: 24,
			Width:      31,
			Height:     1,
			Title:      utils.StrToPtr("Date"),
		})

	// dahsboard widget
	dashboardCtrlr := controller.NewDashboardController(
		&thandler,
		&service,
		&mutex,
		globalChan,
		controller.DashboardProps{
			Width:      80,
			Height:     18,
			RenderPosX: 34,
			RenderPosY: 8,
			Title:      utils.StrToPtr("Dashboard"),
		},
	)

	worklogDescCtrlr := controller.NewWorklogDescController(
		&thandler,
		&service,
		&mutex,
		controller.WorklogDescProps{
			Width:      112,
			Height:     4,
			RenderPosX: 2,
			RenderPosY: 59,
			Title:      utils.StrToPtr("Detail log"),
		},
	)

	worklogCtrlr := controller.NewWorklogController(
		&thandler,
		&service,
		&mutex,
		globalChan,
		2,
		controller.WorklogProps{
			Width:      112,
			Height:     50,
			RenderPosX: 2,
			RenderPosY: 27,
			Title:      utils.StrToPtr("Worklogs"),
		},
		worklogDescCtrlr.ReloadData,
	)

	guideCtrlr := controller.NewGuideController(
		&thandler,
		&mutex,
		controller.GuideProps{
			RenderPosX: 2,
			RenderPosY: 65,
		},
	)

	userCtrlr.ListenFromController()
	userCtrlr.CreateWindow()

	dateCtrlr.ListenFromController()
	dateCtrlr.CreateWindow()

	dashboardCtrlr.ListenFromController()
	dashboardCtrlr.CreateWindow()

	worklogCtrlr.ListenFromController()
	worklogCtrlr.CreateWindow()

	worklogDescCtrlr.ListenFromController()
	worklogDescCtrlr.CreateWindow()

	guideCtrlr.ListenFromController()
	guideCtrlr.CreateWindow()

	ctrlrList := controller.ControllerChild{
		0: userCtrlr.GetChan(),
		1: dateCtrlr.GetChan(),
		2: worklogCtrlr.GetChan(),
		3: worklogDescCtrlr.GetChan(),
		4: dashboardCtrlr.GetChan(),
		5: guideCtrlr.GetChan(),
	}
	ctrl := controller.NewController(
		&wg,
		&thandler,
		&service,
		ctrlrList,
		globalChan,
		userCtrlr.GetSelectedName,
		dateCtrlr.GetDates,
	)

	if err := thandler.Render(); err != nil {
		panic(err)
	}

	wg.Add(1)
	ctrl.SetupController()

	wg.Wait()
}
