package widget

import (
	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type BookList struct {
	w.List
	i int
}

func NewBookList(nodes []BookNode) *BookList {
	l := &BookList{}
	l.List = *w.NewList()
	rows := make([]string, len(nodes))
	for i, node := range nodes {
		rows[i] = node.Title
	}
	l.TextStyle = ui.NewStyle(ui.ColorGreen)
	l.Border = false
	l.WrapText = false
	l.Rows = rows
	return l
}
