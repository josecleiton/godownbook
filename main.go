package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/josecleiton/godownbook/repo"
	"github.com/josecleiton/godownbook/repo/libgen"
	"github.com/josecleiton/godownbook/util"
	w "github.com/josecleiton/godownbook/widget"
)

var searchPattern string
var verboseFlag bool
var repository string

var bookTable *w.BookTable
var pageIndicator *w.PageIndicator

var supportedRepositories = map[string]repo.Repository{
	"libgen": libgen.Make(),
}

func init() {
	flag.StringVar(&searchPattern, "s", "", "book title to search")
	flag.BoolVar(&verboseFlag, "v", false, "verbose log")
	flag.StringVar(&repository, "r", "", "where to lookup book")
	flag.Parse()
}

func test() {
	repo := reposToSearch()
	c, _ := fetchInitialData(repo)
	util.PrintMemUsage()
	br, _ := repo.GetRows(c)
	m, _ := repo.MaxPageNumber(c)
	util.PrintMemUsage()
	log.Println("page", m)
	page, _ := repo.BookInfo(br[0])
	util.PrintMemUsage()
	log.Println(page)
	util.PrintMemUsage()
	log.Println(page.ToBIB())
	os.Exit(0)
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

func setupGrid() *ui.Grid {
	grid := ui.NewGrid()
	grid.Set(ui.NewRow(90.0/100, bookTable), ui.NewRow(10.0/100, pageIndicator))
	return grid
}

func eventLoop() {
	for e := range ui.PollEvents() {
		switch e.ID {
		case "q", "<C-c>", "<C-z>":
			return
		}
	}
}

func main() {
	repo := reposToSearch()
	rows, max := buildRows(repo)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	bookTable = w.NewBookTable(rows)
	pageIndicator = w.NewPageIndicator(max)
	grid := setupGrid()
	tWidth, tHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, tWidth, tHeight)
	ui.Render(grid)

	eventLoop()
}
