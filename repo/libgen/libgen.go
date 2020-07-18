package libgen

import (
	"github.com/josecleiton/godownbook/repo"
)

type LibGen struct {
}

func (lib LibGen) downPage() {
}

func (LibGen) SearchUrl() string {
	return "http://gen.lib.rus.ec/search.php"
}

func (LibGen) QueryField() string {
	return "req"
}

func (LibGen) PaginationField() string {
	return "page"
}

func (LibGen) SortEnabled() bool {
	return true
}

func (LibGen) SortField() string {
	return "sortmode"
}

func (LibGen) SortValues() map[int]string {
	return map[int]string{
		repo.ASC:  "ASC",
		repo.DESC: "DESC",
	}
}

func (LibGen) ExtraFields() map[string]string {
	return map[string]string{
		"phrase": "1",
		"view":   "simple",
		"column": "def",
		"sort":   "def",
	}
}

