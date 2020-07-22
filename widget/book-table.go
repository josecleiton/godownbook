package widget

import (
	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type BookTable struct {
	w.Table
	i int
}

func NewBookTable(rows [][]string) *BookTable {
	table := &BookTable{}
	table.Table = *w.NewTable()
	table.Rows = rows
	table.BorderStyle = ui.NewStyle(ui.ColorGreen)
	table.TextStyle = ui.NewStyle(ui.ColorClear)
	table.TextAlignment = ui.AlignCenter
	table.RowSeparator = true
	table.FillRow = false
	table.RowStyles[0] = ui.NewStyle(ui.ColorMagenta, ui.ColorClear, ui.ModifierBold)
	return table
}
