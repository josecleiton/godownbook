package util

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
)

// FetchBody is a struct to provide body to http requests
type FetchBody struct {
	ContentType string
	Body        string
}

// Fetch make a request to an url
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

// FetchImage get image from url
func FetchImage(url *url.URL) (*image.Image, error) {
	resp, err := Fetch(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func FetchHeader(url *url.URL, header string) (string, error) {
	resp, err := http.Head(url.String())
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return resp.Header.Get(header), nil
}
