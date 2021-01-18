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

	// optimization: pre-calculated fields (used often)
	cellsPerBox int
	cellsPerCol int
	cellsPerRow int
}

func newBoard(boxWidth, boxHeight, boxesWide, boxesTall int) *board {
	var b = &board{
		boxWidth:    boxWidth,
		boxHeight:   boxHeight,
		boxesWide:   boxesWide,
		boxesTall:   boxesTall,
		cellsPerBox: boxWidth * boxHeight,
		cellsPerCol: boxHeight * boxesTall,
		cellsPerRow: boxWidth * boxesWide,
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
		b.checkBoxRepeat,
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

// checkBoxRepeat checks the constraint : No value may be repeated within a box
func (b *board) checkBoxRepeat(duplicates checkerCallback) (err error) {
	var (
		cellIDs = make([]int, b.cellsPerCol)
		offset  int
	)

	for box := 0; box < b.boxesTall*b.boxesWide; box++ {
		offset = b.cellsPerBox * box

		for i := 0; i < b.cellsPerBox; i++ {
			cellIDs[i] = offset + i
		}

		if items := checkDuplicateIDs(b, cellIDs); len(items) > 0 {
			err = fmt.Errorf("box %d contains duplicate values", 1+box)
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
		cellIDs = make([]int, b.cellsPerCol)

		bx, by, i, j, col, colNum, offset int
	)

	for bx, colNum = 0, 0; bx < b.boxesWide; bx++ {
		for col = 0; col < b.boxWidth; col++ {
			i = 0
			for by = 0; by < b.boxesTall; by++ {
				offset = by * b.cellsPerBox * b.boxesWide //+ bx*b.cellsPerBox + col*b.boxWidth

				for j = 0; j < b.boxHeight; j++ {
					cellIDs[i] = offset + j*b.cellsPerRow
					i++
				}
			}

			colNum++
			if items := checkDuplicateIDs(b, cellIDs); len(items) > 0 {
				err = fmt.Errorf("column %d contains duplicate values", colNum)
				if duplicates != nil {
					duplicates(items)
				}
			}
		}
	}

	return err
}

// checkRowRepeat checks the constraint : No value may be repeated within a row
func (b *board) checkRowRepeat(duplicates checkerCallback) (err error) {
	var (
		cellIDs = make([]int, b.cellsPerRow)

		bx, by, i, j, row, rowNum, offset int
	)

	for by, rowNum = 0, 0; by < b.boxesTall; by++ {
		for row = 0; row < b.boxHeight; row++ {
			i = 0
			for bx = 0; bx < b.boxesWide; bx++ {
				offset = by*b.cellsPerBox*b.boxesWide + bx*b.cellsPerBox + row*b.boxWidth

				for j = 0; j < b.boxWidth; j++ {
					cellIDs[i] = offset + j
					i++
				}
			}

			rowNum++
			if items := checkDuplicateIDs(b, cellIDs); len(items) > 0 {
				err = fmt.Errorf("row %d contains duplicate values", rowNum)
				if duplicates != nil {
					duplicates(items)
				}
			}
		}
	}

	return err
}

func checkDuplicateIDs(b *board, ids []int) []*cell {
	var (
		occur = make(map[string][]int)
		value string
	)

	for _, c := range ids {
		if value = b.cells[c].Given; value == "" {
			if value = b.cells[c].Center; value == "" {
				continue
			}
		}

		if _, exist := occur[value]; !exist {
			occur[value] = []int{}
		}

		occur[value] = append(occur[value], c)
	}

	dupes := make([]*cell, 0)
	for _, o := range occur {
		if len(o) > 1 {
			for _, c := range o {
				dupes = append(dupes, b.cells[c])
			}
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

	b.cells = make([]*cell, b.cellsPerBox*numBoxes)

	n := 0
	for i := 0; i < numBoxes; i++ {
		cells := make([]fyne.CanvasObject, b.cellsPerBox)

		for j := 0; j < b.cellsPerBox; j++ {
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
