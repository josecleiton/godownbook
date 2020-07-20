package libgen

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/josecleiton/godownbook/repo"
	"golang.org/x/net/html"
)

const (
	BOOKS_PER_PAGE = 25
)

const (
	id = iota
	author
	title
	publisher
	year
	pages
	language
	filesize
	extension
)

type LibGen struct {
	queryField      string
	baseURL         *url.URL
	paginationField string
	sortEnabled     bool
	sortField       string
	sortModeField   string
	sortModeValues  map[repo.SortMode]string
	columns         []string
	extraFields     map[string]string
}

func New() LibGen {
	u, _ := url.Parse("http://gen.lib.rus.ec/search.php")
	return LibGen{
		queryField:      "req",
		baseURL:         u,
		paginationField: "page",
		sortEnabled:     true,
		sortField:       "sort",
		columns:         []string{"Author", "Title", "Publisher", "Year", "Pages", "Language", "Filesize", "Extension"},
		sortModeField:   "sortmode",
		sortModeValues: map[repo.SortMode]string{
			repo.ASC:  "ASC",
			repo.DESC: "DESC",
		},
		extraFields: map[string]string{
			"phrase": "1",
			"view":   "simple",
			"column": "def",
			"sort":   "def",
		},
	}
}

func (LibGen) HttpMethod() string {
	return http.MethodGet
}

func (l LibGen) BaseURL() *url.URL {
	return l.baseURL
}

func (l LibGen) QueryField() string {
	return l.queryField
}

func (l LibGen) PaginationField() string {
	return l.paginationField
}

func (l LibGen) SortEnabled() bool {
	return l.sortEnabled
}

func (l LibGen) SortField() string {
	return l.sortField
}

func (l LibGen) Columns() []string {
	return l.columns
}

func (l LibGen) SortModeField() string {
	return l.sortModeField
}

func (l LibGen) SortModeValues() map[repo.SortMode]string {
	return l.sortModeValues
}

func (l LibGen) ExtraFields() map[string]string {
	return l.extraFields
}

func (LibGen) ContentType() string {
	return ""
}

func bodyCrowler(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "body" {
		return node, nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n, _ := bodyCrowler(child); n != nil {
			return n, nil
		}
	}
	return nil, errors.New("<body> not found")
}

func bookTableCrowler(node *html.Node) (*html.Node, error) {
	const TABLE_IDX = 3
	i := 0
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "table" {
			if i+1 == TABLE_IDX {
				return child, nil
			}
			i++
		}
	}
	return nil, errors.New(fmt.Sprintf("<table> #%d not found", TABLE_IDX))
}

func bookTbodyCrowler(node *html.Node) (*html.Node, error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "tbody" {
			return child, nil
		}
	}
	return nil, errors.New("<tbody> not found")
}

func bookTrCrowler(node *html.Node) ([]*html.Node, error) {
	list := make([]*html.Node, 0, BOOKS_PER_PAGE)
	// ignore the header
	for child := node.FirstChild.NextSibling; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "tr" {
			list = append(list, child)
		}
	}
	if len(list) == 0 {
		return nil, errors.New("none <tr> found")
	}
	return list, nil
}

func bookTitleTextCrowler(node *html.Node) (string, error) {
	if node.Type == html.TextNode {
		return node.Data, nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.TextNode {
			continue
		}
		return child.Data, nil
	}
	return "", errors.New("book title text not found")
}

func bookTitleCrowler(node *html.Node) (text, url string, err error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !(child.Type == html.ElementNode && child.Data == "a") {
			continue
		}
		for _, attr := range child.Attr {
			if !(attr.Key == "href" && strings.HasPrefix(attr.Val, "book")) {
				continue
			}
			text, err := bookTitleTextCrowler(child.FirstChild)
			return text, attr.Val, err
		}
	}
	return "", "", errors.New("book title not found")
}

func bookTdCrowler(td *html.Node) (string, error) {
	for child := td.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			return child.Data, nil
		}
		log.Println(child.Type == html.ElementNode, child.Data)
	}
	return "", errors.New("text node not found")
}

func newBookRow(tr *html.Node, rowLen int) (*repo.BookRow, error) {
	br := &repo.BookRow{Columns: make([]string, 0, rowLen)}
	i := 0
	for child := tr.FirstChild; child != nil && i < extension; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
			log.Println("idx", i)
			// skip ID
			if i == id {
				i++
				continue
			}
			text := ""
			switch i {
			case author:
				a := child.FirstChild
				text = a.FirstChild.Data
			case title:
				t, urlInfo, err := bookTitleCrowler(child)
				br.InfoPage, err = url.Parse(urlInfo)
				if err != nil {
					return nil, err
				}
				text = t
			default:
				log.Println(child.Data)
				t, err := bookTdCrowler(child)
				if err != nil {
					log.Println("text not found at column", i)
				}
				text = t
			}
			br.Columns = append(br.Columns, text)
			i++
		}
	}
	return br, nil
}

func bookRowCrowler(nodes []*html.Node, rowLen int) ([]*repo.BookRow, error) {
	list := make([]*repo.BookRow, 0, BOOKS_PER_PAGE)
	for i := 0; i < BOOKS_PER_PAGE; i++ {
		br, err := newBookRow(nodes[i], rowLen)
		if err != nil {
			handleErr(err)
		}
		list = append(list, br)
	}
	return list, nil
}

func (l LibGen) GetRows(content string) ([]*repo.BookRow, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		handleErr(err)
	}
	body, err := bodyCrowler(doc)
	if err != nil {
		handleErr(err)
	}
	table, err := bookTableCrowler(body)
	if err != nil {
		handleErr(err)
	}
	tbody, err := bookTbodyCrowler(table)
	if err != nil {
		handleErr(err)
	}
	trList, err := bookTrCrowler(tbody)
	if err != nil {
		handleErr(err)
	}
	brs, err := bookRowCrowler(trList, len(l.columns))
	if err != nil {
		handleErr(err)
	}
	return brs, nil
}

func (LibGen) MaxPageNumber(content string) (int, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		handleErr(err)
	}
	body, err := bodyCrowler(doc)
	if err != nil {
		handleErr(err)
	}
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "script" {
			inEl := child.FirstChild
			if strings.Contains(inEl.Data, "Paginator") {
				re := regexp.MustCompile("\\d+")
				raw := re.Find([]byte(inEl.Data))
				return strconv.Atoi(string(raw))
			}
		}
	}
	return -1, errors.New("max page number not found")
}

func handleErr(err error) {
	log.Fatalln(err)

}
