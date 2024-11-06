package termhandler

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

type Termhandler struct {
	screen   *bufio.Writer
	fd       int
	oldState *term.State
}

type Position [2]int

func NewTermHandler() TermhandlerType {
	return &Termhandler{
		screen: bufio.NewWriter(os.Stdout),
	}
}

// EnableRawMode implements TermhandlerType.
func (t *Termhandler) EnableRawMode() {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		log.Fatalf("error enabling raw mode: %v", err)
		return
	}

	t.fd = fd
	t.oldState = oldState
}

// RestoreMode implements TermhandlerType.
func (t *Termhandler) RestoreMode() {
	term.Restore(t.fd, t.oldState)
}

// GetSize implements termhandlerType.
func (t *Termhandler) GetSize() (int, int) {
	fd := int(os.Stdout.Fd())

	width, height, err := term.GetSize(fd)
	if err != nil {
		panic(err)
	}

	return width, height
}

// Render implements termhandlerType.
func (t *Termhandler) Render() error {
	return t.screen.Flush()
}

func (t *Termhandler) Clear() {
	fmt.Fprint(t.screen, "\033[2J")
}

func (t *Termhandler) MoveCursor(pos Position) {
	fmt.Fprintf(t.screen, "\033[%d;%dH", pos[1], pos[0])
}

func (t *Termhandler) ShowCursor() {
	fmt.Fprint(t.screen, "\033[?25h")
}

func (t *Termhandler) HideCursor() {
	fmt.Fprint(t.screen, "\033[?25l")
}

func (t *Termhandler) Draw(str string) {
	fmt.Fprint(t.screen, str)
}
