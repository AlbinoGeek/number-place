package main

import (
	"testing"

	"fyne.io/fyne/test"

	"github.com/stretchr/testify/assert"
)

// tests
var (
	testGridRepeat = `3,3,3,3,55-6---98-7-195----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
	testColRepeat  = `3,3,3,3,53-6---98-7-195----------6-5--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
	testRowRepeat  = `3,3,3,3,53-6---9857-19-----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
)

func TestBoardLoadCheck(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sodoku")
	assert.Errorf(t, board.load(testGridRepeat), "loading repeats in subgrid should have failed")
	assert.Errorf(t, board.load(testColRepeat), "loading repeats in col should have failed")
	assert.Errorf(t, board.load(testRowRepeat), "loading repeats in row should have failed")
}

func TestBoardUndo(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sodoku")

	cell := board.cells[2]
	old := cell.Center

	// It's our responsibility to save a history state if we're setting manually
	board.registerUndo()
	cell.SetCenter("5")

	assert.Errorf(t, board.check(), "checking repeats in subgrid should have failed")
	assert.EqualValues(t, cell.Center, "5", "SetCenter failed to set expected value")

	board.undo()
	assert.EqualValues(t, cell.Center, old, "undo did not restore value")
}

func BenchmarkBoardLoad(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.load(wikipedia)
	}
}

func BenchmarkBoardCheck(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sodoku")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.check()
	}
}

func BenchmarkBoardCheckSubgrids(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sodoku")

	cb := func(cells []*cell) {
		_ = cells
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.checkSubgridRepeat(cb)
	}
}

func BenchmarkBoardCheckCols(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sodoku")

	cb := func(cells []*cell) {
		_ = cells
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.checkColRepeat(cb)
	}
}

func BenchmarkBoardCheckRows(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sodoku")

	cb := func(cells []*cell) {
		_ = cells
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.checkRowRepeat(cb)
	}
}
