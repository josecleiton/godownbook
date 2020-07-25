package widget

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
	"github.com/josecleiton/godownbook/book"
)

const (
	paddingX = 10
	paddingY = 5
)

type BookModal struct {
	ui.Grid
	Data *book.Book
}

func newParagraph() *w.Paragraph {
	p := w.NewParagraph()
	p.Border = false
	return p
}

func newInfoTxt(b *book.Book) string {
	return fmt.Sprintf(`Author: %s
Publisher: %s
Year: %s
Pages: %s
Lang: %s
Filesize: %s
Ext: %s


%s`, b.Author, b.Publisher,
		b.Year, b.Pages, b.Language,
		b.Size, b.Extension, b.Synopsis)
}

// func newInfoGrid(b *book.Book) *ui.Grid {
// 	info := ui.NewGrid()
// 	info.Border = false
// 	author := newParagraph()
// 	author.Text = "Author: " + b.Author
// 	title := newParagraph()
// 	title.Text = "Title: " + b.Title
// 	publisher := newParagraph()
// 	publisher.Text = "Publisher: " + b.Publisher
// 	year := newParagraph()
// 	year.Title = "Year: " + b.Year
// 	pages := newParagraph()
// 	pages.Title = "Pages: " + b.Pages
// 	lang := newParagraph()
// 	lang.Title = "Lang: " + b.Language
// 	size := newParagraph()
// 	size.Title = "Filesize: " + b.Size
// 	ext := newParagraph()
// 	ext.Title = "Ext: " + b.Extension
// 	const ratio = 8 / 10.0
// 	info.Set(
// 		ui.NewRow(ratio, author), ui.NewRow(ratio, title), ui.NewRow(ratio, publisher),
// 		ui.NewRow(ratio, year), ui.NewRow(ratio, pages), ui.NewRow(ratio, lang),
// 		ui.NewRow(ratio, size), ui.NewRow(ratio, ext),
// 	)
// 	return info
// }

func NewBookModal(b *book.Book, tw, th int) *BookModal {
	bm := &BookModal{Data: b}
	bm.Grid = *ui.NewGrid()
	bm.Resize(tw, th)
	content := w.NewParagraph()
	content.Text = newInfoTxt(b)
	content.Title = b.Title
	bar := w.NewParagraph()
	bar.Text = "Press 'd' to download or 'ESC' to exit"
	// img := w.NewImage(*b.Cover)
	// img.SetRect(0, 0, modalw, modalh)
	bm.Set(ui.NewRow(0.8, content), ui.NewRow(0.2, bar))
	// content := newContent(b)
	// content.SetRect(modalw/10, modalh/10, 2*modalw/3, modalh)
	// bm.Set(ui.NewCol(0.33, img), ui.NewCol(0.77, content))
	return bm
}

func (b *BookModal) Resize(tw, th int) {
	modalw, modalh := 2*tw/3, 2*th/3
	b.SetRect(tw/4, th/4, modalw, modalh)
}

