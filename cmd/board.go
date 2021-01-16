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

// ! BADLY NAMED
type checkerCallback func(dupes []*cell)

// ! BADLY NAMED
type checker func(duplicates checkerCallback) error

func (b *board) check() error {
	b.mu.Lock()

	var cb checkerCallback
	if HighlightMistakes {
		cb = b.setMistakes(true)
	}

	errors := make([]error, 0)
	for _, f := range []checker{
		b.checkSubgridRepeat,
		b.checkColRepeat,
		b.checkRowRepeat,
	} {
		if err := f(cb); err != nil {
			errors = append(errors, err)
		}
	}

	b.mu.Unlock()

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (b *board) setMistakes(v bool) func([]*cell) {
	return func(cells []*cell) {
		for _, c := range cells {
			c.SetMistake(v)
		}
	}
}

// checkSubgridRepeat checks the constraint : No value may be repeated within a subgrid
func (b *board) checkSubgridRepeat(duplicates checkerCallback) (err error) {
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
			err = fmt.Errorf("subgrid %d contains duplicate values", 1+sg)
			if duplicates != nil {
				duplicates(items)
			}
		}
	}

	return err
}

// checkColRepeat checks the constraint : No value may be repeated within a column
func (b *board) checkColRepeat(duplicates checkerCallback) (err error) {
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

	for i, col := range data {
		if items := checkDuplicates(col); len(items) > 0 {
			err = fmt.Errorf("column %d contains duplicate values", 1+i)
			if duplicates != nil {
				duplicates(items)
			}
		}
	}

	return err
}

// checkRowRepeat checks the constraint : No value may be repeated within a row
func (b *board) checkRowRepeat(duplicates checkerCallback) (err error) {
	var (
		cellsPerSG  = b.boxWidth * b.boxHeight
		cellsPerCol = b.boxHeight * b.boxesTall
		cellsPerRow = b.boxWidth * b.boxesWide
		data        = make([][]*cell, cellsPerRow)
		offset      int
	)

	for row := 0; row < cellsPerRow; row++ {
		data[row] = make([]*cell, cellsPerCol)
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

			data[row][col] = b.cells[offset+i]
		}
	}

	for i, row := range data {
		if items := checkDuplicates(row); len(items) > 0 {
			err = fmt.Errorf("row %d contains duplicate values", 1+i)
			if duplicates != nil {
				duplicates(items)
			}
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
		boxObjects []fyne.CanvasObject
	)

	// optimization: reduce allocations by re-using previously-created fyne.Container
	if b.Container != nil {
		boxObjects = b.Container.Objects

		if have := len(boxObjects); have > numBoxes || cap(boxObjects) >= numBoxes {
			// len too high, or len too low, but space in capacity
			boxObjects = boxObjects[:numBoxes]
		} else if have < numBoxes {
			// len too low, needs extend
			boxObjects = append(boxObjects, make([]fyne.CanvasObject, numBoxes-have)...)
		}
	} else {
		boxObjects = make([]fyne.CanvasObject, numBoxes)
	}

	b.cells = make([]*cell, boxSize*numBoxes)

	n := 0
	for i := 0; i < numBoxes; i++ {
		cells := make([]fyne.CanvasObject, boxSize)

		for j := 0; j < boxSize; j++ {
			b.cells[n] = newCell(n)
			cells[j] = b.cells[n]
			n++
		}

		if boxObjects[i] != nil {
			// optimization: reduce allocations by re-using previously-created widget.Card
			box := boxObjects[i].(*widget.Card).Content.(*fyne.Container)
			box.Objects = cells
			box.Refresh()
		} else {
			boxObjects[i] = widget.NewCard("", "", fyne.NewContainerWithLayout(
				layout.NewAdaptiveGridLayout(b.boxWidth),
				cells...,
			))
		}
	}

	if b.Container != nil {
		// optimization: reduce allocations by re-using previously created fyne.Container
		b.Container.Objects = boxObjects
		b.Container.Refresh()
	} else {
		b.Container = fyne.NewContainerWithLayout(
			layout.NewAdaptiveGridLayout(b.boxesWide),
			boxObjects...,
		)
	}
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
	b.setMistakes(false)(b.cells)
	b.mu.Unlock()

	b.check()
}
