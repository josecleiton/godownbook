package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
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
var configFormat string

var uiClosed bool
var bookTable *w.BookTable
var bookTree *w.BookTree
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
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}
	flag.StringVar(&configFormat, "cf", "json", "config format [json, yaml]")
	flag.StringVar(&configPath, "c", filepath.Join(cfgDir, "godownbook", "config."+configFormat), "config file path")
	flag.StringVar(&searchPattern, "s", "", "book title to search")
	flag.BoolVar(&verboseFlag, "v", false, "verbose log")
	flag.StringVar(&repository, "r", "", "where to lookup book")
	flag.Parse()
	err = config.UserConfig.Parse(configPath)
	if err != nil {
		log.Println(err)
	}
	if repository == "" {
		repository = config.UserConfig.DefaultRepo
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
	cFile := make(chan *os.File)
	progress := make(chan float64)
	dest := filepath.Join(config.UserConfig.OutDirBib, b.ToPath())
	downloader, err := repo.DownloadBook("Libgen.lc")
	if err != nil {
		log.Fatalln(err)
	}
	go downloader.Exec(b.Mirrors["Libgen.lc"], dest, cFile, progress)
	for {
		select {
		case f := <-cFile:
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
			util.PrintMemUsage()
			os.Exit(0)
		case p := <-progress:
			log.Println(p)
		}
	}
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
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r)
		nodes[i].Childs = make([]string, 0, len(row.Columns)-len(keyBitmap))
		for j, col := range row.Columns {
			if !keyBitmap[j] {
				nodes[i].Childs = append(nodes[i].Childs, columns[j]+": "+col)
			}
		}
	}
	return nodes, max
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
	grid.Set(ui.NewRow(0.9, bookTree), ui.NewRow(0.1, pageIndicator))
	return grid
}

func eventLoop() {
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, os.Interrupt)
	signal.Notify(sigTerm, os.Kill)
	previousKey := ""
	uiEvents := ui.PollEvents()
	l := bookTree
	for {
		select {
		case <-sigTerm:
			return
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
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
				l.ToggleExpand()
			case "G", "<End>":
				l.ScrollBottom()
			case "E":
				l.ExpandAll()
			case "C":
				l.CollapseAll()
			case "<Resize>":
				x, y := ui.TerminalDimensions()
				l.SetRect(0, 0, x, y)
			}

			if previousKey == "g" {
				previousKey = ""
			} else {
				previousKey = e.ID
			}
			ui.Render(grid)
		}

	}
}

func main() {
	test()
	repo := reposToSearch()
	nodes, max := makeTreeData(repo)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer func() { util.PrintMemUsage() }()
	defer func() {
		if !uiClosed {
			ui.Close()
		}
	}()
	bookTree = w.NewBookTree(nodes)
	pageIndicator = w.NewPageIndicator(max)
	grid = setupGrid()
	tWidth, tHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, tWidth, tHeight)
	ui.Render(grid)

	eventLoop()
}
