package repo

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// SortMode search sort mode
type SortMode int
type FetchStep int

const (
	ASC SortMode = iota
	DESC
)

const (
	RowStep FetchStep = iota
	InfoPageStep
	DownloadStep
)

// ContentError generic error on parsing content
var ContentError = errors.New("content parsing error")

// BookRow represents a book row
type BookRow struct {
	InfoPage *url.URL
	Columns  []string
}

// Repository represents a book repository
type Repository interface {
	// HttpMethod returns the http method to fetch content
	HttpMethod(step FetchStep) string
	// SearchUrl returns the base url of repository
	BaseURL() *url.URL
	// QueryField returns the query field of repository. Ex: ?search=value
	QueryField() string
	//  PaginationField returns the page field of repository. Ex: ?page=2
	PaginationField() string
	// SortEnabled returns if repository allow sorting
	SortEnabled() bool
	// SortField returns the sort field param of repository. Ex: ?sort=author
	SortField() string
	// Colums columns from repo
	Columns() []string
	//SortModeField returns the sort mode field of repository. Ex: ?sortmode=ASC
	SortModeField() string
	// SortModeValues returns a map to ascending and descending sort modes
	SortModeValues() map[SortMode]string
	// ExtraFields any extra field to append into http call
	ExtraFields() map[string]string
	// ContentType content type of repository. Highly recommended in POST calls
	ContentType() string
	// GetRows return rows from content
	GetRows(content string) ([]*BookRow, error)
	// MaxPageNumber return max page number from content
	MaxPageNumber(content string) (int, error)
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

// FetchContent use repository httpMethod to pull the content
func FetchContent(r Repository, url string, step FetchStep) (content string, code int, err error) {
	var resp *http.Response

	switch r.HttpMethod(step) {
	case http.MethodPost:
		resp, err = http.Post(url, r.ContentType(), bytes.NewBuffer([]byte{}))
	default:
		resp, err = http.Get(url)
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), code, err
}

