package main

import (
	"fmt"
	"log"
	"strconv"

	// "syscall"

	// ui "github.com/gizak/termui/v3"
	"github.com/josecleiton/godownbook/config"
	"github.com/josecleiton/godownbook/repo"
	w "github.com/josecleiton/godownbook/widget"
)

func buildRows(r repo.Repository) (rows [][]string, max int) {
	rows = make([][]string, r.MaxPerPage()+1)
	rows[0] = r.Columns()
	c, err := fetchInitialData(r)
	if err != nil {
		log.Fatalln(err)
	}
	br, err := r.GetRows(c)
	if err != nil {
		log.Fatalln(err)
	}
	max, err = r.MaxPageNumber(c)
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < r.MaxPerPage(); i++ {
		rows[i+1] = br[i].Columns
	}

	return
}

func makeListData(r repo.Repository) ([]w.BookNode, int) {
	nodes := make([]w.BookNode, r.MaxPerPage())
	c, err := fetchInitialData(r)
	handleError(err)
	br, err := r.GetRows(c)
	handleError(err)
	max, err := r.MaxPageNumber(c)
	handleError(err)
	for i, row := range br {
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r, config.UserConfig.Delimiter[0])
		if i == 1 {
			nodes[i].Title = fmt.Sprintf("[%s](fg:blue)", nodes[i].Title)
		}
	}
	return nodes, max
}

func makeTreeData(r repo.Repository) ([]w.BookNode, int) {
	nodes := make([]w.BookNode, r.MaxPerPage())
	c, err := fetchInitialData(r)
	handleError(err)
	br, err := r.GetRows(c)
	handleError(err)
	max, err := r.MaxPageNumber(c)
	handleError(err)
	keyColumns := r.KeyColumns()
	keyBitmap := make(map[int]bool, len(keyColumns))
	columns := r.Columns()
	for _, idx := range keyColumns {
		keyBitmap[idx] = true
	}
	for i, row := range br {
		nodes[i].Title = strconv.Itoa(i+1) + ". " + row.Key(r, config.UserConfig.Delimiter[0])
		nodes[i].Childs = make([]string, 0, len(row.Columns)-len(keyBitmap))
		for j, col := range row.Columns {
			if !keyBitmap[j] {
				nodes[i].Childs = append(nodes[i].Childs, columns[j]+": "+col)
			}
		}
	}
	return nodes, max
}
