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

	"github.com/josecleiton/godownbook/book"
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
	httpMethods     map[repo.FetchStep]string
}

func Make() LibGen {
	base, _ := url.Parse("http://gen.lib.rus.ec/search.php")
	return LibGen{
		queryField:      "req",
		baseURL:         base,
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
		httpMethods: map[repo.FetchStep]string{
			repo.RowStep:      http.MethodGet,
			repo.InfoPageStep: http.MethodGet,
		},
	}
}

func (l LibGen) HttpMethod(step repo.FetchStep) string {
	return l.httpMethods[step]
}

func (l LibGen) BaseURL() url.URL {
	return *l.baseURL
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

func bodyCrawler(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "body" {
		return node, nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n, _ := bodyCrawler(child); n != nil {
			return n, nil
		}
	}
	return nil, errors.New("<body> not found")
}

func tableCrawler(node *html.Node, nTable int) (*html.Node, error) {
	i := 0
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "table" {
			if i+1 == nTable {
				return child, nil
			}
			i++
		}
	}
	return nil, errors.New(fmt.Sprintf("<table> #%d not found", nTable))
}

func tbodyCrawler(node *html.Node) (*html.Node, error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "tbody" {
			return child, nil
		}
	}
	return nil, errors.New("<tbody> not found")
}

func trListCrawler(node *html.Node, n int) ([]*html.Node, error) {
	list := make([]*html.Node, 0, n)
	// ignore the header
	i := 0
	for child := node.FirstChild.NextSibling; child != nil && i < n; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "tr" {
			list = append(list, child)
			i++
		}
	}
	if i == 0 {
		return nil, errors.New("none <tr> found")
	}
	return list, nil
}

func bookTitleTextCrawler(node *html.Node) (string, error) {
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

func bookTitleCrawler(node *html.Node) (text, url string, err error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if !(child.Type == html.ElementNode && child.Data == "a") {
			continue
		}
		for _, attr := range child.Attr {
			if !(attr.Key == "href" && strings.HasPrefix(attr.Val, "book")) {
				continue
			}
			text, err := bookTitleTextCrawler(child.FirstChild)
			return text, attr.Val, err
		}
	}
	return "", "", errors.New("book title not found")
}

func textCrawler(node *html.Node) (string, error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			return child.Data, nil
		}
	}
	return "", errors.New("text node not found")
}

func newBookRow(tr *html.Node, rowLen int) (*repo.BookRow, error) {
	br := &repo.BookRow{Columns: make([]string, 0, rowLen)}
	i := 0
	for child := tr.FirstChild; child != nil && i < extension; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
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
				t, urlInfo, err := bookTitleCrawler(child)
				br.InfoPage, err = url.Parse(urlInfo)
				if err != nil {
					return nil, err
				}
				text = t
			default:
				t, err := textCrawler(child)
				if err != nil {
					log.Println("libgen: text not found at column", i)
				}
				text = t
			}
			br.Columns = append(br.Columns, text)
			i++
		}
	}
	return br, nil
}

func bookRowCrawler(nodes []*html.Node, rowLen int) ([]*repo.BookRow, error) {
	list := make([]*repo.BookRow, 0, BOOKS_PER_PAGE)
	for i := 0; i < BOOKS_PER_PAGE; i++ {
		br, err := newBookRow(nodes[i], rowLen)
		if err != nil {
			return []*repo.BookRow{}, err
		}
		list = append(list, br)
	}
	return list, nil
}

func (l LibGen) GetRows(content string) ([]*repo.BookRow, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return []*repo.BookRow{}, err
	}
	body, err := bodyCrawler(doc)
	if err != nil {
		return []*repo.BookRow{}, err
	}
	table, err := tableCrawler(body, 3)
	if err != nil {
		return []*repo.BookRow{}, err
	}
	tbody, err := tbodyCrawler(table)
	if err != nil {
		return []*repo.BookRow{}, err
	}
	trList, err := trListCrawler(tbody, BOOKS_PER_PAGE)
	if err != nil {
		return []*repo.BookRow{}, err
	}
	return bookRowCrawler(trList, len(l.columns))
}

func (LibGen) MaxPageNumber(content string) (int, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return -1, err
	}
	body, err := bodyCrawler(doc)
	if err != nil {
		return -1, err
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

func (l LibGen) BookInfo(b *repo.BookRow) (*book.Book, error) {
	// log.Println("book", b.InfoPag)
	u := l.BaseURL()
	u.Path = b.InfoPage.Path
	u.RawQuery = b.InfoPage.RawQuery
	content, code, err := repo.FetchContent(l, &u, repo.InfoPageStep)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, errors.New("libgen: book info status code != 200")
	}
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	body, err := bodyCrawler(doc)
	if err != nil {
		return nil, err
	}
	const nTable = 1
	table, err := tableCrawler(body, nTable)
	if err != nil {
		return nil, err
	}
	tbody, err := tbodyCrawler(table)
	if err != nil {
		return nil, err
	}
	const nTr = 18
	trList, err := trListCrawler(tbody, nTr)
	if err != nil {
		return nil, err
	}
	log.Println(trList, len(trList))
	book, err := bookInfoCrawler(trList, l.baseURL)
	if err != nil {
		return nil, err
	}
	book.URL = &u
	return book, nil
}

func aCrawler(node *html.Node) (*html.Node, error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "a" {
			return child, nil
		}
	}
	return nil, errors.New("libgen: <a> not found")
}

func imgCrawler(node *html.Node) (*html.Node, error) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "img" {
			return child, nil
		}
	}
	return nil, errors.New("libgen: <img> not found")
}

func textCrawlerDeep(node *html.Node) (*html.Node, error) {
	if node.Type == html.TextNode {
		return node, nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n, _ := textCrawlerDeep(child); n != nil {
			return n, nil
		}
	}
	return nil, errors.New("libgen: text node not found")
}

func trCrawlerDeep(node *html.Node) (*html.Node, error) {
	if node.Type == html.ElementNode && node.Data == "tr" {
		return node, nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if n, _ := trCrawlerDeep(child); n != nil {
			return n, nil
		}
	}
	return nil, errors.New("libgen: <tr> node not found")

}

func foundAttrib(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func attribsToMap(node *html.Node) map[string]string {
	attribs := make(map[string]string, 10)
	for _, attr := range node.Attr {
		attribs[attr.Key] = attr.Val
	}
	return attribs
}

func bookInfoCrawlerTdCover(node *html.Node, b *book.Book, base *url.URL) error {
	a, err := aCrawler(node)
	if err != nil {
		return err
	}
	img, err := imgCrawler(a)
	if err != nil {
		return err
	}
	b.Cover = &url.URL{}
	*b.Cover = *base
	b.Cover.Path = foundAttrib(img, "src")
	return nil
}

func bookInfoCrawlerTd(node *html.Node, b *book.Book) error {
	return nil

}

func bookInfoCrawlerTrCover(node *html.Node, b *book.Book, base *url.URL) error {
	var values [2]string
	i := -1
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
			if i == -1 {
				err := bookInfoCrawlerTdCover(child, b, base)
				if err != nil {
					return err
				}
				i++
				continue
			}
			bookInfoCrawlerKeyValue(child, b, &values, i)
			i++
		}
	}
	return nil
}

func bookInfoCrawlerKeyValue(node *html.Node, b *book.Book, values *[2]string, i int) error {
	txtNode, err := textCrawlerDeep(node)
	if err != nil {
		return err
	}
	values[i%2] = strings.TrimSpace(txtNode.Data)
	if i%2 == 1 && values[0] != "" {
		b.Fill(values[0], values[1])
		values[0] = ""
		values[1] = ""
	}
	return nil
}

func bookInfoCrawlerTr(node *html.Node, b *book.Book) {
	var values [2]string
	i := 0
	// log.Println("começoooou")
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
			err := bookInfoCrawlerKeyValue(child, b, &values, i)
			if err != nil {
				log.Println(err)
			}
			i++
		}
	}
}

func bookInfoCrawlerSynopsis(node *html.Node, b *book.Book) error {
	txtNode, err := textCrawlerDeep(node)
	if err != nil {
		return err
	}
	b.Synopsis = strings.TrimSpace(txtNode.Data)
	return nil
}

func bookInfoCrawlerMirrorsTd(node *html.Node, b *book.Book, nMirrors int) error {
	i := 0
	for child := node.FirstChild; child != nil && i < nMirrors; child = child.NextSibling {
		log.Println(child, i, nMirrors)
		if child.Type == html.ElementNode && child.Data == "td" {
			a, err := aCrawler(child)
			if err != nil {
				return err
			}
			attribs := attribsToMap(a)
			href, err := url.Parse(attribs["href"])
			if err != nil {
				return err
			}
			if attribs["title"] != "" {
				b.Mirrors[attribs["title"]] = href
				i++
			}
		}
	}
	if i != nMirrors {
		return errors.New("libgen: mirror not found")
	}
	return nil
}

func bookInfoCrawlerMirrors(node *html.Node, b *book.Book) error {
	txtNode, err := textCrawlerDeep(node)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(txtNode.Data, "Mirrors") {
		return errors.New("libgen: mirrors not found")
	}
	for child := node.FirstChild.NextSibling; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "td" {
			tr, err := trCrawlerDeep(child)
			if err != nil {
				return err
			}
			return bookInfoCrawlerMirrorsTd(tr, b, 2)
		}
	}
	return errors.New("mirror <td> not found")
}

func bookInfoCrawler(trList []*html.Node, base *url.URL) (*book.Book, error) {
	const (
		cover    = 0
		mirrors  = 16
		synopsis = 17
	)
	var err error
	b := book.New()
	for i, tr := range trList {
		switch i {
		case cover:
			err = bookInfoCrawlerTrCover(tr, b, base)
		case mirrors - 2, mirrors - 1:
			break
		case mirrors:
			err = bookInfoCrawlerMirrors(tr, b)
			break
		case synopsis:
			err = bookInfoCrawlerSynopsis(tr, b)
			break

		default:
			bookInfoCrawlerTr(tr, b)
		}
		if err != nil {
			return nil, err
		}
	}
	return b, err
}
