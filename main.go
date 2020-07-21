package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/josecleiton/godownbook/repo"
	"github.com/josecleiton/godownbook/repo/libgen"
	"github.com/josecleiton/godownbook/util"
)

var searchPattern string
var verboseFlag bool
var repository string

var supportedRepositories = map[string]repo.Repository{
	"libgen": libgen.Make(),
}

func init() {
	flag.StringVar(&searchPattern, "s", "", "book title to search")
	flag.BoolVar(&verboseFlag, "v", false, "verbose log")
	flag.StringVar(&repository, "r", "", "where to lookup book")
	flag.Parse()
}

func reposToSearch() []repo.Repository {
	if repository != "" {
		if supportedRepositories[repository] == nil {
			keys := make([]string, 0, len(supportedRepositories))
			for k := range supportedRepositories {
				keys = append(keys, k)
			}
			log.Fatalf("Use a supported repository: [%v]\n", strings.Join(keys, ", "))
		}
		return []repo.Repository{supportedRepositories[repository]}
	}
	repos := make([]repo.Repository, 0, len(supportedRepositories))
	for _, v := range supportedRepositories {
		repos = append(repos, v)
	}
	return repos
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

func test() {
	repos := reposToSearch()
	log.Println(repos)
	c, _ := fetchInitialData(repos[0])
	util.PrintMemUsage()
	br, _ := repos[0].GetRows(c)
	for _, row := range br {
		log.Println(*row)
	}
	m, _ := repos[0].MaxPageNumber(c)
	util.PrintMemUsage()
	log.Println("page", m)
	page, _ := repos[0].BookInfo(br[0])
	util.PrintMemUsage()
	log.Println(page)
	util.PrintMemUsage()
	log.Println(page.ToBIB())
	os.Exit(0)
}

func main() {
	test()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Text = "Hello World!"
	p.SetRect(0, 0, 25, 5)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
