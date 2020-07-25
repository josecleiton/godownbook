package widget

import (
	ui "github.com/gizak/termui/v3"
	// w "github.com/gizak/termui/v3/widgets"
)

type MainScreen struct {
	ui.Grid
	BookList      *BookList
	PageIndicator *PageIndicator
	CPi           chan int
	CBl           chan int
}

func NewMainScreen(bl *BookList, pi *PageIndicator, tw, th int) *MainScreen {
	ms := &MainScreen{
		BookList: bl, PageIndicator: pi,
		CPi: make(chan int), CBl: make(chan int),
	}
	ms.Grid = *ui.NewGrid()
	ms.Set(ui.NewRow(0.9, bl), ui.NewRow(0.1, pi))
	ms.SetRect(0, 0, tw, th)
	return ms
}

func (ms *MainScreen) Update() {
	ms.Set(ui.NewRow(0.9, ms.BookList), ui.NewRow(0.1, ms.PageIndicator))
}

func (ms *MainScreen) Resize() {
	ms.Lock()
	tw, th := ui.TerminalDimensions()
	ms.Unlock()
	ms.SetRect(0, 0, tw, th)
}
