package controller

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	termhandler "tui/term-handler"
)

type GuideProps struct {
	RenderPosX int
	RenderPosY int
}

type GuideController struct {
	handler      termhandler.TermhandlerType
	activeGuide  int
	guideOptions map[int]string
	mutex        *sync.Mutex
	localChan    chan string
	props        GuideProps
}

func NewGuideController(
	handler *termhandler.TermhandlerType,
	mutex *sync.Mutex,
	guideProps GuideProps,
) GuideControllerType {
	return &GuideController{
		handler:     *handler,
		mutex:       mutex,
		localChan:   make(chan string, 2),
		props:       guideProps,
		activeGuide: 0,
		guideOptions: map[int]string{
			0: "[k][j] / [][] : Up Down │ [i] : Search │ [Enter] : Interact │ [q] : Quit",
			1: "[k][j][h][l] / [][][][] : Up Down Left Right │ [Enter] : Interact │ [q] : Quit",
			2: "[k][j][h][l] / [][][][] : Up Down Left Right │ [Enter] : Interact │ [q] : Quit",
			3: "[k][j] / [][] : Up Down │ [Enter] : Back │ [q] : Quit",
		},
	}
}

// GetChan implements GuideControllerType.
func (g *GuideController) GetChan() chan<- string {
	return g.localChan
}

// ListenFromController implements GuideControllerType.
func (g *GuideController) ListenFromController() {
	go func() {
		for resChan := range g.localChan {
			switch resChan {
			case "1", "2", "3", "4":
				aw, err := strconv.Atoi(resChan)
				if err != nil {
					continue
				}

				g.mutex.Lock()
				g.activeGuide = aw - 1
				g.cleanBody()
				g.renderGuide()
				g.mutex.Unlock()

			case Resize:
				g.mutex.Lock()
				g.CreateWindow()
				g.mutex.Unlock()
			}
		}
	}()
}

// CreateWindow implements GuideControllerType.
func (g *GuideController) CreateWindow() {
	g.handler.MoveCursor(
		termhandler.Position{
			1,
			1,
		},
	)

	banner := `
    ▗▖ ▗▖▗▄▖▗▄▄▖▗▖ ▗▗▖   ▗▄▖ ▗▄▖▗▄▄▄      ▗▄▖▗▖  ▗▖▗▄▖▗▖▗▖  ▗▗▄▄▗▄▄▄▖▗▄▄▖
    ▐▌ ▐▐▌ ▐▐▌ ▐▐▌▗▞▐▌  ▐▌ ▐▐▌ ▐▐▌  █    ▐▌ ▐▐▛▚▖▐▐▌ ▐▐▌ ▝▚▞▐▌    █ ▐▌   
    ▐▌ ▐▐▌ ▐▐▛▀▚▐▛▚▖▐▌  ▐▌ ▐▐▛▀▜▐▌  █    ▐▛▀▜▐▌ ▝▜▐▛▀▜▐▌  ▐▌ ▝▀▚▖ █  ▝▀▚▖
    ▐▙█▟▝▚▄▞▐▌ ▐▐▌ ▐▐▙▄▄▝▚▄▞▐▌ ▐▐▙▄▄▀    ▐▌ ▐▐▌  ▐▐▌ ▐▐▙▄▄▐▌▗▄▄▞▗▄█▄▗▄▄▞▘
  `
	g.handler.Draw(fmt.Sprintf("\033[36;1m%s\033[0m", banner))
	g.renderGuide()
}

func (g *GuideController) renderGuide() {
	g.handler.MoveCursor(
		termhandler.Position{
			g.props.RenderPosX,
			g.props.RenderPosY,
		},
	)

	guideText, _ := g.guideOptions[g.activeGuide]
	widgetsOptionsText := ""
	if g.activeGuide != 3 {
		widgetsOptionsText = "[1][2][3] : Change Widget │ "
	}
	g.handler.Draw(fmt.Sprintf("\033[32;1m%s%s\033[0m", widgetsOptionsText, guideText))
	g.handler.Render()
}

func (g *GuideController) cleanBody() {
	g.handler.MoveCursor(
		termhandler.Position{
			g.props.RenderPosX,
			g.props.RenderPosY,
		},
	)

	g.handler.Draw(fmt.Sprintf("%s", strings.Repeat(" ", 100)))
	g.handler.Render()
}
