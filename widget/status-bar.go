package widget

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type StatusBar struct {
	ui.Grid
	download      *w.Paragraph
	title         *w.Paragraph
	msg           *w.Paragraph
	downCount     int
	finishedCount int
	progress      int
}

func NewStatusBar() *StatusBar {
	s := &StatusBar{
		title:    w.NewParagraph(),
		msg:      w.NewParagraph(),
		download: w.NewParagraph(),
	}
	s.Grid = *ui.NewGrid()
	s.title.Text = "godownbook"
	s.title.Border = false
	s.download.Border = false
	s.msg.Border = false
	s.Update()
	return s
}

func (s *StatusBar) Update() {
	s.Set(ui.NewCol(0.2, s.title), ui.NewCol(0.6, s.msg), ui.NewCol(0.2, s.download))
}

func (s *StatusBar) OnDownload() int {
	s.Lock()
	s.downCount++
	s.updateDownText()
	s.Unlock()
	return s.downCount
}

func (s *StatusBar) OnError() int {
	s.Lock()
	s.downCount--
	s.updateDownText()
	s.Unlock()
	return s.downCount
}

func (s *StatusBar) OnMessage(msg string) {
	s.Lock()
	s.msg.Text = msg
	s.Unlock()
}

func (s *StatusBar) OnFinished() int {
	s.Lock()
	s.finishedCount++
	s.downCount--
	s.progress = 0
	s.updateDownText()
	s.Unlock()
	return s.finishedCount
}

func (s *StatusBar) OnProgress(percent float64) int {
	s.Lock()
	s.progress = int(percent * 100)
	s.updateDownText()
	s.Unlock()
	return s.progress
}

func (s *StatusBar) updateDownText() {
	var text, progress string
	// 🔽: downloading
	// ✅: done
	if s.downCount > 0 {
		text = fmt.Sprintf("🔽 %d", s.downCount)
		if s.downCount == 1 {
			progress = fmt.Sprintf("%d", s.progress) + "% "
		}
	}
	if s.finishedCount > 0 {
		text = fmt.Sprintf("✅ %d %s", s.finishedCount, text)
	}
	s.download.Text = progress + text
}
