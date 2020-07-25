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
	bm.Set(ui.NewRow(0.75, content), ui.NewRow(0.25, bar))
	// content := newContent(b)
	// content.SetRect(modalw/10, modalh/10, 2*modalw/3, modalh)
	// bm.Set(ui.NewCol(0.33, img), ui.NewCol(0.77, content))
	return bm
}

func (b *BookModal) Resize(tw, th int) {
	modalw, modalh := 2*tw/3, 2*th/3
	b.SetRect(tw/4, th/4, modalw, modalh)
}

