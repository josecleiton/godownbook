package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	// "syscall"

	ui "github.com/gizak/termui/v3"
	"github.com/josecleiton/godownbook/book"
	"github.com/josecleiton/godownbook/config"
	"github.com/josecleiton/godownbook/repo"
	w "github.com/josecleiton/godownbook/widget"
)

type BookController struct {
	Display  chan *w.BookModal
	Download chan int
}

func NewBookController() *BookController {
	return &BookController{
		Display:  make(chan *w.BookModal),
		Download: make(chan int),
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func makeListData(r repo.Repository, br []*repo.BookRow) []w.BookNode {
	nodes := make([]w.BookNode, len(br))
	for i, row := range br {
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r, config.UserConfig.Delimiter[0])
		if i == 1 {
			nodes[i].Title = fmt.Sprintf("[%s](fg:blue)", nodes[i].Title)
		}
	}
	return nodes
}

func fetchBookRows(r repo.Repository, queryOpts *repo.QueryOptions, step repo.FetchStep) ([]*repo.BookRow, int) {
	c, err := repo.FetchData(r, queryOpts, step)
	handleError(err)
	br, err := r.GetRows(c)
	handleError(err)
	max, err := r.MaxPageNumber(c)
	handleError(err)
	return br, max
}

func terminalDim() (int, int) {
	wRender.Lock()
	tw, th := ui.TerminalDimensions()
	wRender.Unlock()
	return tw, th
}

func fetchInitialData(r repo.Repository, load chan int) ([]*repo.BookRow, int) {
	load <- 33
	queryOpts := repo.NewQueryOptions(searchPattern)
	br, max := fetchBookRows(r, queryOpts, repo.RowStep)
	load <- 66
	return br, max
}

func fetchData(r repo.Repository, load chan int, done chan bool) {
	defer func() { done <- true }()
	br, max := fetchInitialData(r, load)
	nodes := makeListData(r, br)
	time.Sleep(50 * time.Millisecond)
	tw, th := terminalDim()
	mainScreen := w.NewMainScreen(w.NewBookList(nodes), w.NewPageIndicator(max, 1), tw, th)
	load <- LOAD_COMPLETED
	iDone := make(chan bool)
	bc := NewBookController()
	go eventLoop(mainScreen, bc, iDone)
	var book *book.Book
	var err error
	for {
		select {
		case <-iDone:
			return
		case selectedRow := <-mainScreen.SelectedRow:
			book, err = r.BookInfo(br[selectedRow])
			if err != nil {
				bc.Display <- nil
				break
			}
			tw, th := terminalDim()
			bc.Display <- w.NewBookModal(book, tw, th)
		case mirrorIdx := <-bc.Download:
			if mirrorIdx < 0 {
				break
			}
			log.Println("LANÇOU")
		case page := <-mainScreen.UpdatePage:
			log.Println(page)
		}
	}
}
