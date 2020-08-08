package widget

import (
	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type BookList struct {
	w.List
	highlighted bool
}

func NewBookList(nodes []BookNode) *BookList {
	l := &BookList{}
	l.List = *w.NewList()
	rows := make([]string, len(nodes))
	for i, node := range nodes {
		rows[i] = node.Title
	}
	l.TextStyle = ui.NewStyle(ui.ColorGreen)
	l.Border = true
	l.WrapText = false
	l.Rows = rows
	return l
}

func (l *BookList) ToggleHighlight() {
	l.highlighted = !l.highlighted
	l.drawHighlight()
}

func (l *BookList) drawHighlight() {
	if l.highlighted {
		l.BorderStyle = ui.NewStyle(ui.ColorBlue)
	} else {
		l.BorderStyle = ui.Theme.Block.Border
	}
}
