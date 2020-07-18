package repo

const (
	ASC = iota
	DESC
)

type Repository interface {
	MaxPageNumber(page string, n int) int
	Name() string
	Alias() string
}
