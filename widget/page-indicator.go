package widget

import (
	// "errors"
	"strconv"

	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type PageIndicator struct {
	w.TabPane
	Selected    int
	max         int
	highlighted bool
}

func NewPageIndicator(max int) *PageIndicator {
	pages := make([]string, max)
	for i := range pages {
		pages[i] = strconv.Itoa(i + 1)
	}
	pi := &PageIndicator{TabPane: *w.NewTabPane(pages...), max: max}
	pi.Border = true
	return pi
}

func (pi *PageIndicator) FocusStart() {
	pi.ActiveTabIndex = 0
}

func (pi *PageIndicator) FocusEnd() {
	pi.ActiveTabIndex = pi.max - 1
}

func (pi *PageIndicator) FocusMiddle() {
	pi.ActiveTabIndex = pi.max / 2
}

func (pi *PageIndicator) ToggleHighlight() {
	pi.highlighted = !pi.highlighted
	pi.drawHighlight()
}

func (pi *PageIndicator) drawHighlight() {
	if pi.highlighted {
		pi.BorderStyle = ui.NewStyle(ui.ColorGreen)
	} else {
		pi.BorderStyle = ui.Theme.Block.Border
	}
}

// func (p *PageIndicator) SelectPage(page int) error {
// 	if page > p.Max {
// 		return errors.New("page selected is out of bound")
// 	}
// 	p.selected = page
// 	return nil
// }

// func (p *PageIndicator) SelectedPage() int {
// 	return p.selected
// }
