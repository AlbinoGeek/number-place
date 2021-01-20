package main

import (
	"fmt"
	"strconv"
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

// This should be a perfectly valid (but not solvable) grid that looks like:
//
// ABC JKL
// DEF MNO
// GHI PQR
var testBoardKey = `3,3,2,1,ABCDEFGHIJKLMNOPQR`

func TestUIBoardKey(t *testing.T) {
	var (
		a        = test.NewApp()
		board, w = start(a)
	)
	a.Run()

	test.AssertImageMatches(t, "start-empty.png", w.Canvas().Capture())

	assert.NoError(t, board.load(testBoardKey), "failed loading test board key")

	v := board.Container
	_ = v

	test.AssertImageMatches(t, "board-key.png", w.Canvas().Capture())
}

func TestUICells(t *testing.T) {
	var (
		a        = test.NewApp()
		board, w = start(a)
	)
	a.Run()

	test.AssertImageMatches(t, "start-empty.png", w.Canvas().Capture())

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")

	test.AssertImageMatches(t, "start.png", w.Canvas().Capture())

	// Test all given cells
	c := board.cells[0]
	w.SetContent(c)
	for i := 1; i < 10; i++ {
		c.SetGiven(strconv.Itoa(i))
		c.Refresh()

		test.AssertImageMatches(t, fmt.Sprintf("cell-given-%d.png", i), w.Canvas().Capture())
	}

	// Test empty cells
	c = board.cells[2]
	w.SetContent(c)
	c.Refresh()

	test.AssertImageMatches(t, "cell-empty.png", w.Canvas().Capture())

	// Test center-valued cells
	for i := 1; i < 10; i++ {
		c.SetCenter(strconv.Itoa(i))
		c.Refresh()

		test.AssertImageMatches(t, fmt.Sprintf("cell-center-%d.png", i), w.Canvas().Capture())
	}
}

func TestUICellSelect(t *testing.T) {
	var (
		a        = test.NewApp()
		board, w = start(a)

		cells = []*cell{
			board.cells[0],
			board.cells[13],
			board.cells[26],
		}
	)
	a.Run()

	// test setSelected
	for _, c := range cells {
		c.Select()
	}

	v := "1"
	setSelected(board, v)()
	test.AssertImageMatches(t, "some-set.png", w.Canvas().Capture())

	for _, c := range cells {
		assert.Equal(t, v, c.Center)
		assert.Equal(t, false, c.mistake)
		c.Select()
	}

	test.AssertImageMatches(t, "some-set-selected.png", w.Canvas().Capture())

	// test clearSelected
	clearSelected(board)()
	v = ""
	for _, c := range cells {
		assert.Equal(t, v, c.Center)
	}
}
