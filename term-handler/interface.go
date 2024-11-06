package termhandler

type TermhandlerType interface {
  Clear()
  GetSize() (int, int)
  Draw(str string)
  Render() error
  MoveCursor(place Position)
  ShowCursor()
  HideCursor()
  EnableRawMode()
  RestoreMode()
}
