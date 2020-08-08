package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/josecleiton/godownbook/config"
	"github.com/josecleiton/godownbook/repo"
	"github.com/josecleiton/godownbook/repo/libgen"
	"github.com/josecleiton/godownbook/util"
	w "github.com/josecleiton/godownbook/widget"
	tb "github.com/nsf/termbox-go"
)

var searchPattern string
var verboseFlag bool
var repository string
var configPath string

var wRender = &sync.Mutex{}

const (
	LOAD_COMPLETED = 100
)

var supportedRepositories = map[string]repo.Repository{
	"libgen": libgen.Make(),
}

func init() {
	err := config.Init()
	if err != nil {
		log.Fatalln(err)
	}
	ucdir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}
	cfgdir := filepath.Join(ucdir, "godownbook")
	flag.StringVar(&configPath, "c", filepath.Join(cfgdir, "config.json"), "config file path")
	flag.StringVar(&searchPattern, "s", "", "book title to search")
	flag.BoolVar(&verboseFlag, "v", false, "verbose log")
	flag.StringVar(&repository, "r", "", "where to lookup book")
	flag.Parse()
	parseConfigFile(cfgdir)
	if repository == "" {
		repository = config.UserConfig.DefaultRepo
	}
}

func parseConfigFile(cdir string) {
	exts := [2]string{config.JSON, config.YAML}
	for _, ext := range exts {
		fp := "config." + ext
		log.Printf("trying config file \"%s\"", fp)
		if err := config.UserConfig.Parse(filepath.Join(cdir, fp)); err == nil {
			log.Printf("config file \"%s\" loaded", fp)
			return
		}
	}
}

func lockAndRender(items ...ui.Drawable) {
	wRender.Lock()
	ui.Render(items...)
	wRender.Unlock()
}

func handleResize(items ...w.Resizable) {
	wRender.Lock()
	tw, th := ui.TerminalDimensions()
	for _, item := range items {
		item.Resize(tw, th)
	}
	wRender.Unlock()
}

func toggleHighlight(items ...w.Highlightable) {
	for _, item := range items {
		item.ToggleHighlight()
	}
}

func screenResize(grid *ui.Grid) {
	tw, th := ui.TerminalDimensions()
	grid.SetRect(0, 0, tw, th)
}

func reposToSearch() repo.Repository {
	if supportedRepositories[repository] == nil {
		keys := make([]string, 0, len(supportedRepositories))
		for k := range supportedRepositories {
			keys = append(keys, k)
		}
		log.Fatalf("Use a supported repository: [%v]\n", strings.Join(keys, ", "))
	}
	return supportedRepositories[repository]
}

type PageType int

// TODO: split this func in other funcs
func eventLoop(mainScreen *w.MainScreen, bc *BookController, done chan bool) {
	const (
		LIST PageType = iota
		MODAL
		PAGES
	)
	defer func() { done <- true }()
	var modal *w.BookModal
	l := mainScreen.BookList
	highlighted := LIST
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, os.Interrupt)
	signal.Notify(sigTerm, os.Kill)
	previousKey := ""
	wRender.Lock()
	uiEvents := ui.PollEvents()
	ui.Render(mainScreen)
	wRender.Unlock()
	for {
		select {
		case <-sigTerm:
			return
		case modal = <-bc.Display:
			if modal != nil {
				highlighted = MODAL
				lockAndRender(modal)
			}
		case percentage := <-mainScreen.UpdateDown:
			mainScreen.StatusBar.OnProgress(percentage)
			lockAndRender(mainScreen)
		case f := <-mainScreen.DownloadedFile:
			f.Close()
			mainScreen.StatusBar.OnMessage(filepath.Base(f.Name()) + " downloaded")
			mainScreen.StatusBar.OnFinished()
			lockAndRender(mainScreen)
			mainScreen.StatusBar.OnMessage("")
		case <-mainScreen.SelectedRow:
		case e := <-uiEvents:
			l = mainScreen.BookList
			// global key maps
			switch e.ID {
			case "q", "<C-c>":
				return
			}
			if highlighted == LIST {
				switch e.ID {
				case "d", "D":
					return
				case "j", "<Down>":
					l.ScrollDown()
				case "k", "<Up>":
					l.ScrollUp()
				case "<C-d>", "L":
					l.ScrollHalfPageDown()
				case "<C-u>", "H":
					l.ScrollHalfPageUp()
				case "<C-f>", "J":
					l.ScrollPageDown()
				case "<C-b>", "K":
					l.ScrollPageUp()
				case "g":
					if previousKey == "g" {
						l.ScrollTop()
					}
				case "<Home>":
					l.ScrollTop()
				case "<Enter>":
					mainScreen.SelectedRow <- l.SelectedRow
				case "G", "<End>":
					l.ScrollBottom()
				case "<Resize>":
					handleResize(mainScreen)
				case "<Tab>", "p", "P":
					toggleHighlight(mainScreen.PageIndicator, mainScreen.BookList)
					highlighted = PAGES
				}
				if num, err := strconv.Atoi(e.ID); (num > 0 || previousKey != "") && err == nil {
					if num2, err := strconv.Atoi(previousKey); err == nil {
						num = num2*10 + num
					}
					previousKey = e.ID
					if num > 9 {
						previousKey = ""
						if num > 25 {
							num = 25
						}
					}
					l.SelectedRow = num - 1
				} else {
					if previousKey == "g" {
						previousKey = ""
					} else {
						previousKey = e.ID
					}
				}
				lockAndRender(mainScreen)
			} else if highlighted == MODAL {
				switch e.ID {
				case "d", "<Enter>", "<Space>":
					bc.Download <- "Libgen.lc"
					fallthrough
				case "<Escape>", "c", "C":
					highlighted = LIST
					lockAndRender(mainScreen)
				case "<Resize>":
					handleResize(modal)
					lockAndRender(modal)
				}
			} else { //highlighted ==PAGES
				pi := mainScreen.PageIndicator
				switch e.ID {
				case "<Tab>", "b", "B":
					toggleHighlight(pi, mainScreen.BookList)
					highlighted = LIST
					if pi.Selected != pi.ActiveTabIndex {
						pi.ActiveTabIndex = pi.Selected
					}
				case "h":
					pi.FocusLeft()
				case "l":
					pi.FocusRight()
				case "<Enter>":
					pi.Selected = pi.ActiveTabIndex
					// load page
				}
				lockAndRender(mainScreen)
			}
		}

	}
}

func loadingWidget() *w.Loading {
	lw := w.NewLoading()
	tw, th := ui.TerminalDimensions()
	if tw > 50 {
		tw = 50
	}
	if th > 5 {
		th = 5
	}
	lw.SetRect(0, 0, tw, th)
	return lw
}

func updatePercentage(lw *w.Loading, cstat chan int) {
	for {
		percent := <-cstat
		lw.Percent = percent
		lockAndRender(lw)
		if lw.Percent == LOAD_COMPLETED {
			return
		}
	}
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	tb.SetInputMode(tb.InputEsc) // disable mouse input
	defer func() { util.PrintMemUsage() }()
	defer ui.Close()
	r := reposToSearch()
	lw := loadingWidget()
	ui.Render(lw)
	loadProgress := make(chan int)
	done := make(chan bool)
	go updatePercentage(lw, loadProgress)
	go fetchData(r, loadProgress, done)
	<-done
}
