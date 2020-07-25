package widget

import (
	"os"

	ui "github.com/gizak/termui/v3"
	// w "github.com/gizak/termui/v3/widgets"
)

type MainScreen struct {
	ui.Grid
	BookList       *BookList
	PageIndicator  *PageIndicator
	StatusBar      *StatusBar
	UpdateList     chan *BookList
	UpdatePage     chan int
	SelectedRow    chan int
	UpdateDown     chan float64
	DownloadedFile chan *os.File
}

func NewMainScreen(sb *StatusBar, bl *BookList, pi *PageIndicator, tw, th int) *MainScreen {
	ms := &MainScreen{
		StatusBar: sb, BookList: bl, PageIndicator: pi,
		UpdatePage: make(chan int), UpdateList: make(chan *BookList),
		UpdateDown: make(chan float64), SelectedRow: make(chan int),
		DownloadedFile: make(chan *os.File),
	}
	ms.Grid = *ui.NewGrid()
	ms.Set(ui.NewRow(0.1, sb), ui.NewRow(0.8, bl), ui.NewRow(0.1, pi))
	ms.SetRect(0, 0, tw, th)
	return ms
}

func (ms *MainScreen) Update() {
	ms.Set(ui.NewRow(0.9, ms.BookList), ui.NewRow(0.1, ms.PageIndicator))
}

func (ms *MainScreen) Resize(tw, th int) {
	ms.SetRect(0, 0, tw, th)
}

