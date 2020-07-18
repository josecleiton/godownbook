package libgen

import (
	"github.com/josecleiton/godownbook/repo"
)

type LibGen struct {
	queryField      string
	searchUrl       string
	paginationField string
	sortEnabled     bool
	sortField       string
	sortValues      map[int]string
	extraFields     map[string]string
}

func Init() LibGen {
	return LibGen{
		queryField:      "req",
		searchUrl:       "http://gen.lib.rus.ec/search.php",
		paginationField: "page",
		sortEnabled:     true,
		sortField:       "sortmode",
		sortValues: map[int]string{
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

func (l LibGen) SortValues() map[int]string {
	return l.sortValues
}

func (l LibGen) ExtraFields() map[string]string {
	return l.extraFields
}

