package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/golog"
)

type state struct {
	Data []byte
	Name string
	Time int64
}

type board struct {
	*fyne.Container `json:"-"`

	mu      sync.Mutex
	cells   []*cell
	history []*state

	boxWidth  int
	boxHeight int
	boxesWide int
	boxesTall int
}

func newBoard(boxWidth, boxHeight, boxesWide, boxesTall int) *board {
	var b = &board{
		boxWidth:  boxWidth,
		boxHeight: boxHeight,
		boxesWide: boxesWide,
		boxesTall: boxesTall,
	}

	b.init()
	return b
}

func (b *board) check() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// constraint : No value may be repeated within a subgrid
	if err := b.checkSubgridRepeat(); err != nil {
		return err
	}

	return b.checkColRepeat()
}

func (b *board) checkSubgridRepeat() (err error) {
	var (
		cellsPerSG = b.boxHeight * b.boxWidth
		grid       = make([]*cell, cellsPerSG)
		offset     int
	)

	for sg := 0; sg < b.boxesTall*b.boxesWide; sg++ {
		offset = cellsPerSG * sg

		// per cell in subgrid
		for i := 0; i < cellsPerSG; i++ {
			grid[i] = b.cells[offset+i]
		}

		if items := checkDuplicates(grid); len(items) > 0 {
			for _, c := range items {
				c.SetMistake(true)
			}

			err = fmt.Errorf("duplicates in subgroup")
		}
	}

	return err
}

func (b *board) checkColRepeat() (err error) {
	var (
		cellsPerSG  = b.boxWidth * b.boxHeight
		cellsPerCol = b.boxHeight * b.boxesTall
		cellsPerRow = b.boxWidth * b.boxesWide
		data        = make([][]*cell, cellsPerCol)
		offset      int
	)

	for col := 0; col < cellsPerRow; col++ {
		data[col] = make([]*cell, cellsPerRow)
	}

	// for each subgrid in board
	for sg := 0; sg < b.boxesTall*b.boxesWide; sg++ {
		offset = cellsPerSG * sg

		// for each cell in subgrid
		for i := 0; i < cellsPerSG; i++ {
			// column
			col := i%b.boxWidth + (sg*b.boxWidth)%(b.boxWidth*b.boxesWide)
			// row
			row := (i / b.boxWidth) + (sg/b.boxesWide)*b.boxesTall

			data[col][row] = b.cells[offset+i]
		}
	}

	for _, row := range data {
		if items := checkDuplicates(row); len(items) > 0 {
			for _, c := range items {
				c.SetMistake(true)
			}

			err = fmt.Errorf("duplicates in row")
		}
	}

	return err
}

func checkDuplicates(data []*cell) []*cell {
	var (
		occur = make(map[string][]*cell)
		value string
	)

	for _, c := range data {
		if value = c.Given; value == "" {
			if value = c.Center; value == "" {
				continue
			}
		}

		if _, exist := occur[value]; !exist {
			occur[value] = []*cell{}
		}

		occur[value] = append(occur[value], c)
	}

	dupes := make([]*cell, 0)
	for _, o := range occur {
		if len(o) > 1 {
			dupes = append(dupes, o...)
		}
	}

	return dupes
}

func (b *board) init() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.history = make([]*state, 0)

	var (
		// TODO: support other cell arrangements, counts, in an elegant way
		boxSize    = b.boxWidth * b.boxHeight
		numBoxes   = b.boxesWide * b.boxesTall
		boxObjects = make([]fyne.CanvasObject, numBoxes)
	)

	b.cells = make([]*cell, boxSize*numBoxes)

	n := 0
	for i := 0; i < numBoxes; i++ {
		cells := make([]fyne.CanvasObject, boxSize)

		for j := 0; j < boxSize; j++ {
			b.cells[n] = newCell(n)
			cells[j] = b.cells[n]
			n++
		}

		boxObjects[i] = widget.NewCard("", "", fyne.NewContainerWithLayout(
			layout.NewAdaptiveGridLayout(b.boxWidth),
			cells...,
		))
	}

	b.Container = fyne.NewContainerWithLayout(
		layout.NewAdaptiveGridLayout(b.boxesWide),
		boxObjects...,
	)
}

func (b *board) load(in string) error {
	parts := strings.Split(in, ",")

	if len(parts) != 5 {
		return errors.New("bad format")
	}

	p, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("bad boxWidth should be integer")
	}
	b.boxWidth = p

	if p, err = strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("bad boxHeight should be integer")
	}
	b.boxHeight = p

	if p, err = strconv.Atoi(parts[2]); err != nil {
		return fmt.Errorf("bad boxesWide should be integer")
	}
	b.boxesWide = p

	if p, err = strconv.Atoi(parts[3]); err != nil {
		return fmt.Errorf("bad boxesTall should be integer")
	}
	b.boxesTall = p

	b.init()

	if a, b := len(parts[4]), len(b.cells); a != b {
		return fmt.Errorf("bad data has wrong cell count: expected %d, got %d", a, b)
	}

	b.mu.Lock()
	for i, c := range b.cells {
		if v := parts[4][i]; v != '-' {
			c.SetGiven(string(v))
		}
	}
	b.mu.Unlock()

	b.registerUndo()
	return b.check()
}

func (b *board) registerUndo() {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := jsoniter.Marshal(&b.cells)
	if err != nil {
		golog.Fatalf("unable to save state: %v", err)
	}

	b.history = append(b.history, &state{
		Data: data,
		Name: "",
		Time: time.Now().Unix(),
	})
}

func (b *board) undo() {
	if len(b.history) == 0 {
		return
	}

	b.mu.Lock()

	n := len(b.history) - 1
	s := b.history[n]
	b.history = b.history[:n]

	jsoniter.Unmarshal(s.Data, &b.cells)
	for _, c := range b.cells {
		c.mistake = false
		c.Refresh()
	}
	b.mu.Unlock()

	b.check()
}
