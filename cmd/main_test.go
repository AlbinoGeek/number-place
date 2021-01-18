package main

import (
	"fmt"
	"strconv"
	"testing"

	"fyne.io/fyne/test"

	"github.com/stretchr/testify/assert"
)

func TestUICells(t *testing.T) {
	var (
		a     = test.NewApp()
		w     = a.NewWindow("Number Place")
		board = newBoard(3, 3, 3, 3)
	)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")

	uiInit(board, w)

	w.Show()
	a.Run()

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
