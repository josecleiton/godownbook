package repo

import (
	"log"
	"net/url"
	"strconv"
)

type SortMode int

const (
	ASC SortMode = iota
	DESC
)

type Repository interface {
	HttpMethod() string
	SearchUrl() string
	QueryField() string
	PaginationField() string
	SortEnabled() bool
	SortField() string
	SortValues() []string
	SortModeField() string
	SortModeValues() map[SortMode]string
	ExtraFields() map[string]string
}

// BaseUrl return url base search url from repository
func BaseUrl(r Repository) *url.URL {
	url, err := url.Parse(r.SearchUrl())
	if err != nil {
		log.Fatalln(err)
	}
	return url
}

// QueryPage appends pagination field to url params
func QueryPage(r Repository, params *url.Values, page int) {
	params.Add(r.PaginationField(), strconv.Itoa(page))
}

// Query appends main search pattern to url params
func Query(r Repository, params *url.Values, value string) {
	params.Add(r.QueryField(), value)
}

// QuerySort appends sort field and sort modifier to url params
func QuerySort(r Repository, params *url.Values, value string, mode SortMode) {
	params.Add(r.SortField(), value)
	if modeField := r.SortModeField(); modeField != "" {
		params.Add(modeField, r.SortModeValues()[mode])
	}
}

// QueryExtraFields appends any extra fields to url params
func QueryExtraFields(r Repository, params *url.Values) {
	for k, v := range r.ExtraFields() {
		params.Add(k, v)
	}
}

