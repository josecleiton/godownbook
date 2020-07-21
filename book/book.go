package book

import (
	"net/url"
	"strings"
)

const (
	title     = "Title"
	author    = "Author"
	publisher = "Publisher"
	isbn      = "ISBN"
	id        = "ID"
	size      = "Size"
	year      = "Year"
	edition   = "Edition"
	extension = "Extension"
)

type Book struct {
	Title     string
	ID        string
	Author    string
	Publisher string
	ISBN      string
	Year      string
	Series    string
	Size      string
	Extension string
	Edition   string
	Volume    string
	URL       *url.URL
	Language  string
	Cover     *url.URL
	Synopsis  string
	Pages     string
	Mirrors   map[string]*url.URL
	ExtraInfo map[string]string
}

func New() *Book {
	return &Book{
		Mirrors:   map[string]*url.URL{},
		ExtraInfo: map[string]string{},
	}
}

func (b Book) ToBIB() string {
	return ""
}

func (b *Book) Fill(key string, value string) {
	if strings.HasPrefix(key, title) {
		b.Title = value
	} else if strings.HasPrefix(key, author) {
		b.Author = value
	} else if strings.HasPrefix(key, publisher) {
		b.Publisher = value
	} else if strings.HasPrefix(key, isbn) {
		b.ISBN = value
	} else if strings.HasPrefix(key, id) {
		b.ID = value
	} else if strings.HasPrefix(key, size) {
		b.Size = value
	} else if strings.HasPrefix(key, year) {
		b.Year = value
	} else if strings.HasPrefix(key, edition) {
		b.Edition = value
	} else if strings.HasPrefix(key, extension) {
		b.Extension = value
	} else if value != "" {
		b.ExtraInfo[key] = value
	}
}
