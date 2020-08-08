package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	Download chan string
}

func NewBookController() *BookController {
	return &BookController{
		Display:  make(chan *w.BookModal),
		Download: make(chan string),
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

func downloadBook(
	downloader repo.Downloader, b *book.Book,
	cfile chan *os.File, cprogress chan float64,
) {
	mirror := downloader.Key()
	dest := filepath.Join(config.UserConfig.OutDir, b.ToPath())
	f, err := downloader.Exec(b.Mirrors[mirror], dest, cfile, cprogress)
	if err != nil {
		return
	}
	f.Close()
	bibPath := filepath.Join(config.UserConfig.OutDirBib, b.ToPathBIB())
	bibFile, err := os.Create(bibPath)
	if err != nil {
		return
	}
	defer bibFile.Close()
	if _, err = bibFile.WriteString(b.ToBIB()); err != nil {
		return
	}
	if userCmd := config.UserConfig.ExecCmd; userCmd != "" {
		cmd := exec.Command(userCmd, f.Name(), bibFile.Name())
		if err := cmd.Start(); err != nil {
			log.Fatalln(err)
		}
	}
}

func fetchData(r repo.Repository, load chan int, done chan bool) {
	defer func() { done <- true }()
	br, max := fetchInitialData(r, load)
	page := 1
	cache := make(map[int][]*repo.BookRow, max)
	cache[page] = br
	nodes := makeListData(r, br)
	time.Sleep(50 * time.Millisecond)
	tw, th := terminalDim()
	mainScreen := w.NewMainScreen(
		w.NewStatusBar(), w.NewBookList(nodes), w.NewPageIndicator(max),
		tw, th,
	)
	load <- LOAD_COMPLETED
	iDone := make(chan bool)
	bc := NewBookController()
	go eventLoop(mainScreen, bc, iDone)
	var book *book.Book
	for {
		select {
		case <-iDone:
			return
		case selectedRow := <-mainScreen.SelectedRow:
			book, err := r.BookInfo(br[selectedRow])
			if err != nil {
				bc.Display <- nil
				break
			}
			tw, th := terminalDim()
			bc.Display <- w.NewBookModal(book, tw, th)
		case mirror := <-bc.Download:
			if downloader, err := r.DownloadBook(mirror); err == nil {
				log.Println("TESTE2")
				mainScreen.StatusBar.OnDownload()
				go downloadBook(downloader, book, mainScreen.DownloadedFile, mainScreen.UpdateDown)
			}
		case page = <-mainScreen.UpdatePage:
			if cache[page] != nil {
			}
			log.Println(page)
		}
	}
}
