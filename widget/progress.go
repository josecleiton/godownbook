package widget

import (

	// ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type Loading struct {
	w.Gauge
}

func NewLoading(cpercent chan int) *Loading {
	lw := &Loading{}
	lw.Gauge = *w.NewGauge()
	lw.Title = "Loading"
	return lw
}

