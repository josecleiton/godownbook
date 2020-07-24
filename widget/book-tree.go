package widget

import (
	ui "github.com/gizak/termui/v3"
	w "github.com/gizak/termui/v3/widgets"
)

type BookTree struct {
	w.Tree
	i int
}

type nodeValue string

func (nv nodeValue) String() string {
	return string(nv)
}

func NewBookTree(nodes []BookNode) *BookTree {
	tree := &BookTree{}
	tree.Tree = *w.NewTree()
	wNodes := make([]*w.TreeNode, len(nodes))
	for i, node := range nodes {
		childs := make([]*w.TreeNode, len(node.Childs))
		wNodes[i] = &w.TreeNode{Value: nodeValue(node.Title), Nodes: childs}
		for j, v := range node.Childs {
			childs[j] = &w.TreeNode{Value: nodeValue(v)}
		}
	}
	tree.TextStyle = ui.NewStyle(ui.ColorGreen)
	tree.Border = false
	tree.SetNodes(wNodes)
	return tree
}
