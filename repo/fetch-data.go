package repo

import (
	"errors"
	"net/url"
)

type QueryOptions struct {
	Page     int
	Sort     string
	SortMode SortMode
	Search   string
}

func NewQueryOptions(search string) *QueryOptions {
	return &QueryOptions{Search: search, Page: 1}
}

func FetchData(r Repository, q *QueryOptions, step FetchStep) (string, error) {
	u := r.BaseURL()
	params := &url.Values{}
	Query(r, params, q.Search)
	if q.Sort != "" {
		QuerySort(r, params, q.Sort, q.SortMode)
	}
	QueryExtraFields(r, params)
	u.RawQuery = params.Encode()
	content, code, err := FetchContent(r, &u, step)
	if code/100 != 2 {
		err = errors.New("fetch data status code != 2xx")
	}
	return content, err
}

