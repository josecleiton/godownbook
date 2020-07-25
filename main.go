package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	// "syscall"

	ui "github.com/gizak/termui/v3"
	"github.com/josecleiton/godownbook/config"
	"github.com/josecleiton/godownbook/repo"
	"github.com/josecleiton/godownbook/repo/libgen"
	"github.com/josecleiton/godownbook/util"
	w "github.com/josecleiton/godownbook/widget"
)

var searchPattern string
var verboseFlag bool
var repository string
var configPath string

var wRender = &sync.Mutex{}

var cdownloading = make(chan bool)

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

func test() {
	repo := reposToSearch()
	c, _ := fetchInitialData(repo)
	util.PrintMemUsage()
	br, _ := repo.GetRows(c)
	m, _ := repo.MaxPageNumber(c)
	util.PrintMemUsage()
	log.Println("page", m)
	b, _ := repo.BookInfo(br[0])
	util.PrintMemUsage()
	log.Println(b)
	util.PrintMemUsage()
	log.Println(b.ToBIB())
	cf := make(chan *os.File)
	progress := make(chan float64)
	dest := filepath.Join(config.UserConfig.OutDirBib, b.ToPath())
	downloader, err := repo.DownloadBook("Libgen.lc")
	if err != nil {
		log.Fatalln(err)
	}
	util.PrintMemUsage()
	go downloader.Exec(b.Mirrors["Libgen.lc"], dest, cf, progress)
	for {
		select {
		case f := <-cf:
			util.PrintMemUsage()
			if f == nil {
				log.Fatalln(errors.New("download fail"))
			}
			defer f.Close()
			log.Println(f.Name())
			log.Println("dowloaded", b.ToPath(), b.Title)
			bib, err := os.Create(filepath.Join(config.UserConfig.OutDirBib, b.ToPathBIB()))
			if err != nil {
				log.Fatalln(err)
			}
			defer bib.Close()
			if _, err = bib.WriteString(b.ToBIB()); err != nil {
				log.Fatalln(err)
			}
			if userCmd := config.UserConfig.ExecCmd; userCmd != "" {
				cmd := exec.Command(userCmd, f.Name(), bib.Name())
				if err := cmd.Start(); err != nil {
					log.Fatalln(err)
				}
			}
			util.PrintMemUsage()
			os.Exit(0)
		case p := <-progress:
			log.Println(p)
		}
	}
}

func lockAndRender(items ...ui.Drawable) {
	wRender.Lock()
	ui.Render(items...)
	wRender.Unlock()
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

func fetchInitialData(r repo.Repository) (content string, err error) {
	u := r.BaseURL()
	log.Println("baseUrl", u)
	params := &url.Values{}
	repo.Query(r, params, searchPattern)
	repo.QueryExtraFields(r, params)
	log.Println("params", params)
	u.RawQuery = params.Encode()
	log.Println("url", u)
	content, code, err := repo.FetchContent(r, &u, repo.RowStep)
	if code != 200 {
		err = errors.New("fetch initial data status code != 200")
	}
	return
}

type PageType int

func eventLoop(mainScreen *w.MainScreen, controller *BookController, done chan bool) {
	const (
		LIST PageType = iota
		MODAL
	)
	defer func() { done <- true }()
	l := mainScreen.BookList
	page := LIST
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
		case modal := <-controller.Display:
			if modal != nil {
				page = MODAL
				lockAndRender(modal)
			}
		case <-sigTerm:
			return
		case e := <-uiEvents:
			l = mainScreen.BookList
			switch e.ID {
			case "q", "<C-c>":
				return
			}
			if page == LIST {
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
					mainScreen.CPi <- l.SelectedRow
				case "G", "<End>":
					l.ScrollBottom()
				case "<Resize>":
					mainScreen.Resize()
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
			}
		}

	}
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	defer func() { util.PrintMemUsage() }()
	r := reposToSearch()
	loadProgress := make(chan int)
	done := make(chan bool)
	lw := w.NewLoading(loadProgress)
	lw.SetRect(0, 0, 50, 5)
	ui.Render(lw)
	go func() {
		for {
			percent := <-loadProgress
			lw.Percent = percent
			lockAndRender(lw)
			if lw.Percent >= 100 {
				break
			}
		}
		loadProgress <- 100
	}()
	go fetchData(r, loadProgress, done)
	<-done
}
