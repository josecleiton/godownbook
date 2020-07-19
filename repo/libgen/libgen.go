package libgen

import (
	"net/http"

	"github.com/josecleiton/godownbook/repo"
)

type LibGen struct {
	queryField      string
	searchUrl       string
	paginationField string
	sortEnabled     bool
	sortField       string
	sortValues      []string
	sortModeField   string
	sortModeValues  map[repo.SortMode]string
	extraFields     map[string]string
}

func NewLibGen() LibGen {
	return LibGen{
		queryField:      "req",
		searchUrl:       "http://gen.lib.rus.ec/search.php",
		paginationField: "page",
		sortEnabled:     true,
		sortField:       "sort",
		sortValues:      []string{"id", "author", "title", "publisher", "year", "pages", "language", "filesize", "extension"},
		sortModeField:   "sortmode",
		sortModeValues: map[repo.SortMode]string{
			repo.ASC:  "ASC",
			repo.DESC: "DESC",
		},
		extraFields: map[string]string{
			"phrase": "1",
			"view":   "simple",
			"column": "def",
			"sort":   "def",
		},
	}
}

func (lib LibGen) downPage() {
}

func (LibGen) HttpMethod() string {
	return http.MethodGet
}

func (l LibGen) SearchUrl() string {
	return l.searchUrl
}

func (l LibGen) QueryField() string {
	return l.queryField
}

func (l LibGen) PaginationField() string {
	return l.paginationField
}

func (l LibGen) SortEnabled() bool {
	return l.sortEnabled
}

func (l LibGen) SortField() string {
	return l.sortField
}

func (l LibGen) SortValues() []string {
	return l.sortValues
}

func (l LibGen) SortModeField() string {
	return l.sortModeField
}

func (l LibGen) SortModeValues() map[repo.SortMode]string {
	return l.sortModeValues
}

func (l LibGen) ExtraFields() map[string]string {
	return l.extraFields
}

func (LibGen) ContentType() string {
	return ""
}

func (LibGen) GetRows(content string, n int) [][]string {
	return [][]string{}
}

func (LibGen) MaxPageNumber(content string, n int) int {
	return 42
}

