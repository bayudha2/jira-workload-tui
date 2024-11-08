package controller

type GuideControllerType interface {
	GetChan() chan<- string
	CreateWindow()
	renderGuide()
	cleanBody()
	ListenFromController()
}

type WindowControllerType interface {
	GetActiveCursor() int
	GetOffsite() int
	GetChan() chan<- string
	GetSelectedName() string
	CreateWindow()
	renderReload()
	cleanBody()
	renderList()
	renderBodyList(*string, int, string)
	renderSearch()
	reloadActiveIndicator()
	ListenFromController()
	handleSearch()
}

type ControllerType interface {
	ListenKeyPress()
	SetupController()
	listenExit()
	listenResize()
	listenRelistenKeyPress()
	exitApp()
}

type DateControllerType interface {
	GetChan() chan<- string
	GetDates() (int, int)
	CreateWindow()
	ListenFromController()
	renderBody()
	reloadActiveIndicator()
}

type DashboardControllerType interface {
	GetChan() chan<- string
	CreateWindow()
	cleanBody()
	renderBody()
	generateRGBChart(int) string
	ListenFromController()
}

type WorklogControllerType interface {
	GetChan() chan<- string
	GetDateCursor() int
	CreateWindow()
	renderBody()
	renderReload()
	reloadActiveIndicator()
  reloadActiveDateIndicator()
	calculateTimespentHighlight(seconds int) string
	mapWorklogData()
	ListenFromController()
}

type WorklogDescControllerType interface {
	GetChan() chan<- string
	ReloadData(dateCursor int)
	CreateWindow()
	renderBody()
	cleanBody()
	ListenFromController()
}
