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
	Download chan bool
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func buildRows(r repo.Repository) (rows [][]string, max int) {
	rows = make([][]string, r.MaxPerPage()+1)
	rows[0] = r.Columns()
	c, err := fetchInitialData(r)
	if err != nil {
		log.Fatalln(err)
	}
	br, err := r.GetRows(c)
	if err != nil {
		log.Fatalln(err)
	}
	max, err = r.MaxPageNumber(c)
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < r.MaxPerPage(); i++ {
		rows[i+1] = br[i].Columns
	}

	return
}

func makeListData(r repo.Repository, br []*repo.BookRow, n int) []w.BookNode {
	nodes := make([]w.BookNode, n)
	for i, row := range br {
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r, config.UserConfig.Delimiter[0])
		if i == 1 {
			nodes[i].Title = fmt.Sprintf("[%s](fg:blue)", nodes[i].Title)
		}
	}
	return nodes
}

func makeTreeData(r repo.Repository) ([]w.BookNode, int) {
	nodes := make([]w.BookNode, r.MaxPerPage())
	c, err := fetchInitialData(r)
	handleError(err)
	br, err := r.GetRows(c)
	handleError(err)
	max, err := r.MaxPageNumber(c)
	handleError(err)
	keyColumns := r.KeyColumns()
	keyBitmap := make(map[int]bool, len(keyColumns))
	columns := r.Columns()
	for _, idx := range keyColumns {
		keyBitmap[idx] = true
	}
	for i, row := range br {
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r, config.UserConfig.Delimiter[0])
		nodes[i].Childs = make([]string, 0, len(row.Columns)-len(keyBitmap))
		for j, col := range row.Columns {
			if !keyBitmap[j] {
				nodes[i].Childs = append(nodes[i].Childs, columns[j]+": "+col)
			}
		}
	}
	return nodes, max
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

func fetchData(r repo.Repository, load chan int, done chan bool) {
	defer func() { done <- true }()
	load <- 33
	queryOpts := repo.NewQueryOptions(searchPattern)
	br, max := fetchBookRows(r, queryOpts, repo.RowStep)
	load <- 66
	nodes := makeListData(r, br, max)
	time.Sleep(50 * time.Millisecond)
	tw, th := terminalDim()
	mainScreen := w.NewMainScreen(w.NewBookList(nodes), w.NewPageIndicator(max, 1), tw, th)
	load <- 100
	<-load
	iDone := make(chan bool)
	bc := &BookController{
		Display:  make(chan *w.BookModal),
		Download: make(chan bool),
	}
	go eventLoop(mainScreen, bc, iDone)
	var book *book.Book
	var err error
	for {
		select {
		case <-iDone:
			return
		case selectedRow := <-mainScreen.CPi:
			book, err = r.BookInfo(br[selectedRow])
			if err != nil {
				bc.Display <- nil
				break
			}
			tw, th := terminalDim()
			bc.Display <- w.NewBookModal(book, tw, th)
		case download := <-bc.Download:
			if !download {
				break
			}
			log.Println("LANÇOU")
		case page := <-mainScreen.CBl:
			log.Println(page)
		}
	}
}
