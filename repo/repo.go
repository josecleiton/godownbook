package repo

const (
	ASC = iota
	DESC
)

type Repository interface {
	SearchUrl() string
	QueryField() string
	PaginationField() string
	SortEnabled() bool
	SortField() string
	SortValues() map[int]string
	ExtraFields() map[string]string
}
