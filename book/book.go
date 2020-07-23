package book

import (
	"fmt"
	"image"
	"net/url"
	"regexp"
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
	page      = "Pages"
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
	Cover     *image.Image
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
	} else if strings.HasPrefix(key, page) {
		b.Pages = value
	} else if value != "" {
		b.ExtraInfo[key] = value
	}
}

func (b Book) ToBIB() string {
	var url string
	if b.URL != nil {
		url = b.URL.String()
	}
	return fmt.Sprintf(`@book{book:%s,
title        =    {%s},
author       =    {%s},
publisher    =    {%s},
isbn         =    {%s},
year         =    {%s},
series       =    {%s},
edition      =    {%s},
volume       =    {%s},
url          =    {%s},
}`, b.ID, b.Title, b.Author, b.Publisher, b.ISBN, b.Year, b.Series, b.Edition, b.Volume, url)
}

func (b Book) titlePath() (t string) {
	re := regexp.MustCompile("[[:punct:]]|_|[[:space:]]+")
	for i, s := range re.Split(b.Title, -1) {
		if s != "" {
			if i != 0 {
				t += "-"
			}
			t += s
		}
	}
	return
}

func (b Book) ToPath() string {
	p := b.titlePath()
	return strings.ToLower(p + "." + b.Extension)
}

func (b Book) ToPathBIB() string {
	p := strings.ToLower(b.titlePath())
	return p + ".bib"
}
