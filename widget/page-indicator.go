package widget

import (
	"errors"
	"strconv"

	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type PageIndicator struct {
	w.Table
	Max      int
	selected int
}

func NewPageIndicator(max int) *PageIndicator {
	pi := &PageIndicator{Max: max}
	pi.Table = *w.NewTable()
	pages := make([]string, max)
	for i := range pages {
		pages[i] = strconv.Itoa(i + 1)
	}
	pi.Rows = [][]string{pages}
	pi.RowSeparator = false
	pi.TextAlignment = ui.AlignCenter
	pi.Border = false
	return pi
}

func (p *PageIndicator) SelectPage(page int) error {
	if page > p.Max {
		return errors.New("page selected is out of bound")
	}
	p.selected = page
	return nil
}

func (p *PageIndicator) SelectedPage() int {
	return p.selected
}
