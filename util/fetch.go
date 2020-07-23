package util

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type FetchBody struct {
	ContentType string
	Body        string
}

func Fetch(u *url.URL, method string, b *FetchBody) (resp *http.Response, err error) {
	url := u.String()
	switch method {
	case http.MethodGet:
		resp, err = http.Get(url)
	case http.MethodPost:
		resp, err = http.Post(url, b.ContentType, strings.NewReader(b.Body))
	default:
		err = errors.New("method not allowed")
	}
	return
}

