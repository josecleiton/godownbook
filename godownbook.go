package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/josecleiton/godownbook/repo"
	"github.com/josecleiton/godownbook/repo/libgen"
)

var searchPattern string
var verboseFlag bool
var repository string

var supportedRepositories = map[string]repo.Repository{
	"libgen": libgen.NewLibGen(),
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

func downPage(r repo.Repository) {
	switch r.HttpMethod() {
	case http.MethodGet:
		log.Println("get")
	}
}

func main() {
	repos := reposToSearch()
	log.Println(repos)
	os.Exit(0)
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
