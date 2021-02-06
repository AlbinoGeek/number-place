package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	gameTimer       *canvas.Text
	gameDuration    binding.String

	solved     binding.Bool
	timeStart  time.Time
	timeFinish time.Time

	mu      sync.Mutex
	cells   []*cell
	history []*state
	initial *state

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
		gameTimer:    canvas.NewText("", theme.ForegroundColor()),
		gameDuration: binding.NewString(),
		boxWidth:     boxWidth,
		boxHeight:    boxHeight,
		boxesWide:    boxesWide,
		boxesTall:    boxesTall,
		cellsPerBox:  boxWidth * boxHeight,
		cellsPerCol:  boxHeight * boxesTall,
		cellsPerRow:  boxWidth * boxesWide,
		solved:       binding.NewBool(),
	}

	b.gameTimer.Hide()

	timeFormat := func(t time.Duration) string {
		return fmt.Sprintf("%.1fs", t.Seconds())
	}

	b.gameDuration.AddListener(binding.NewDataListener(func() {
		last, _ := b.gameDuration.Get()
		end := time.Now()
		if b.timeFinish.UnixNano() > b.timeStart.UnixNano() {
			end = b.timeFinish
		}

		// ! recursive data binding refresh, is this safe/sane?
		// ? maybe this should be triggered by animation instead
		go func() {
			time.Sleep(time.Millisecond * 125)
			if s := timeFormat(end.Sub(b.timeStart)); s != last {
				b.gameTimer.Text = s
				b.gameTimer.Refresh()

				b.gameDuration.Set(s)
			}
		}()
	}))

	b.solved.AddListener(binding.NewDataListener(func() {
		solved, _ := b.solved.Get()
		b.mu.Lock()
		for _, c := range b.cells {
			if solved {
				c.Disable()
			} else {
				c.Enable()
			}
		}
		b.mu.Unlock()
	}))

	b.init()
	return b
}

// UndoIndex is used by the undo function to determine the target state.
type UndoIndex int

var (
	// InitialState will trigger undo to return the board to its' initial state,
	// that is, what the board looked like after it was last loaded.
	InitialState UndoIndex = -2

	// RecentState will trigger undo to return the board before the last change
	// had taken place. This is akin to a traditional undo feature.
	RecentState UndoIndex = -1
)

func (b *board) Reset() {
	if b.initial != nil {
		b.undo(-2)
	}
}

func (b *board) Solved() bool {
	if solved, _ := b.solved.Get(); solved {
		return true
	}

	if b.check() != nil {
		return false
	}

	for _, c := range b.cells {
		if c.Given == "" && c.Center == "" {
			return false
		}
	}

	b.timeFinish = time.Now()
	b.solved.Set(true)
	return true
}

func (b *board) Undo() {
	b.undo(RecentState)
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
		cellIDs = make([]int, b.cellsPerBox)
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
				offset = by*b.cellsPerBox*b.boxesWide + bx*b.cellsPerBox + col

				for j = 0; j < b.boxHeight; j++ {
					cellIDs[i] = offset + j*b.boxWidth
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

func (b *board) getBox(box int) (cells []int) {
	cells = make([]int, b.cellsPerBox)

	boxy := box / b.boxesWide
	boxx := box - boxy*b.boxesWide
	offset := boxy*b.cellsPerRow*b.boxHeight + boxx*b.boxWidth

	for i, y := 0, 0; y < b.boxHeight; y++ {
		for x := 0; x < b.boxWidth; x++ {
			cells[i] = y*b.cellsPerRow + x + offset
			i++
		}
	}


	return
}

func (b *board) init() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.history = make([]*state, 0)
	b.timeStart = time.Now()

	b.solved.Set(false)

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
				layout.NewGridLayout(b.boxWidth),
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
			layout.NewGridLayout(b.boxesWide),
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

	b.cellsPerBox = b.boxWidth * b.boxHeight
	b.cellsPerCol = b.boxHeight * b.boxesTall
	b.cellsPerRow = b.boxWidth * b.boxesWide

	b.init()

	if a, b := len(parts[4]), len(b.cells); a != b {
		return fmt.Errorf("bad data has wrong cell count: expected %d, got %d", b, a)
	}

	b.mu.Lock()
	for i, c := range b.cells {
		if v := parts[4][i]; v != '-' {
			c.SetGiven(string(v))
		}
	}
	b.mu.Unlock()

	b.registerUndo()
	b.initial = b.history[0]

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

func (b *board) undo(idx UndoIndex) {
	if len(b.history) == 0 {
		return
	}

	b.mu.Lock()

	mod := int(idx)
	if l := len(b.history); idx == RecentState || mod > l {
		mod = l - 1
	}

	var s *state

	if idx == InitialState {
		s = b.initial
	} else {
		s = b.history[mod]
		b.history = b.history[:mod]
	}

	jsoniter.Unmarshal(s.Data, &b.cells)
	b.setMistakes(false)(b.cells)
	b.mu.Unlock()

	b.check()
}
