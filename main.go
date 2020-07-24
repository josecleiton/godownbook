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

var uiClosed bool
var bRows []*repo.BookRow
var bookList *w.BookList
var pageIndicator *w.PageIndicator
var grid *ui.Grid

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

func screenResize() {
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

func handleError(err error) {
	if grid != nil {
		ui.Close()
		uiClosed = true
	}
	if err != nil {
		log.Fatalln(err)
	}
}

func setupGrid() *ui.Grid {
	grid := ui.NewGrid()
	grid.Set(ui.NewRow(0.9, bookList), ui.NewRow(0.1, pageIndicator))
	return grid
}

func eventLoop() {
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, os.Interrupt)
	signal.Notify(sigTerm, os.Kill)
	previousKey := ""
	uiEvents := ui.PollEvents()
	l := bookList
	for {
		select {
		case <-sigTerm:
			return
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "d", "D":
				return
			case "j", "<Down>":
				l.ScrollDown()
			case "k", "<Up>":
				l.ScrollUp()
			case "<C-d>":
				l.ScrollHalfPageDown()
			case "<C-u>":
				l.ScrollHalfPageUp()
			case "<C-f>":
				l.ScrollPageDown()
			case "<C-b>":
				l.ScrollPageUp()
			case "g":
				if previousKey == "g" {
					l.ScrollTop()
				}
			case "<Home>":
				l.ScrollTop()
			case "<Enter>":
				break
			case "G", "<End>":
				l.ScrollBottom()
			case "<Resize>":
				screenResize()
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
			ui.Render(grid)
		}

	}
}

func main() {
	// test()
	repo := reposToSearch()
	nodes, max := makeListData(repo)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer func() { util.PrintMemUsage() }()
	defer func() {
		if !uiClosed {
			ui.Close()
		}
	}()
	bookList = w.NewBookList(nodes)
	pageIndicator = w.NewPageIndicator(max)
	grid = setupGrid()
	screenResize()
	ui.Render(grid)

	eventLoop()
}
