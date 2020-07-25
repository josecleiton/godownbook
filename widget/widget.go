package widget

type BookNode struct {
	Title  string
	Childs []string
}

type Resizable interface {
	Resize(x, y int)
}

