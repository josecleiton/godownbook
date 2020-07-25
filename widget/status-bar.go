package widget

import (
	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type StatusBar struct {
	ui.Grid
}

func NewStatusBar() *StatusBar {
	sb := &StatusBar{}
	sb.Grid = *ui.NewGrid()
	title := w.NewParagraph()
	title.Border = false
	title.Text = "godownbook"
	downIndicator := w.NewParagraph()
	downIndicator.Border = false
	downIndicator.Text = " 0 0B/s"
	sb.Set(ui.NewCol(0.8, title), ui.NewCol(0.2, downIndicator))
	return sb
}

