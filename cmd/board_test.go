package main

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

// tests
var (
	testBoxRepeat = `3,3,3,3,55-6---98-7-195----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
	testColRepeat = `3,3,3,3,53-6---98-7-195----------6-5--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
	testRowRepeat = `3,3,3,3,53-6---9857-19-----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
	testEasySolve = `2,2,2,2,-234341223414123`
)

func TestBoardLoadCheck(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")
	assert.Errorf(t, board.load(testBoxRepeat), "loading repeats in boxes should have failed")
	assert.Errorf(t, board.load(testColRepeat), "loading repeats in col should have failed")
	assert.Errorf(t, board.load(testRowRepeat), "loading repeats in row should have failed")
}

func TestBoardSolved(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(2, 2, 2, 2)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")

	assert.Equal(t, false, board.Solved(), "puzzle with blanks should not be solved")

	assert.Errorf(t, board.load(testBoxRepeat), "loading repeats in boxes should have failed")

	assert.Equal(t, false, board.Solved(), "invalid puzzle should not be solved")

	assert.NoError(t, board.load(testEasySolve), "failed loading easy solve")

	assert.Equal(t, false, board.Solved(), "easy solve should not be pre-solved")

	// insure positive solve time
	time.Sleep(time.Millisecond)

	board.cells[0].SetCenter("1")

	assert.Equal(t, true, board.Solved(), "easy solve should now be solved")

	assert.Greater(t, board.timeFinish.UnixNano(), board.timeStart.UnixNano(), "finish time should be  positive")
}

func TestBoardInitExpand(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(2, 2, 2, 2)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")
}

func TestBoardUndo(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sudoku")

	board.registerUndo()
	state := board.history[0]

	cell := board.cells[2]
	old := cell.Center

	// It's our responsibility to save a history state if we're setting manually
	board.registerUndo()
	cell.SetCenter("5")

	assert.Errorf(t, board.check(), "checking repeats in boxes should have failed")
	assert.EqualValues(t, cell.Center, "5", "SetCenter failed to set expected value")

	board.Undo()
	assert.EqualValues(t, cell.Center, old, "undo did not restore value")

	// Set many values
	for _, i := range []int{9, 18, 14, 27, 31, 2, 51, 60, 73} {
		v := strconv.Itoa(rand.Intn(9))

		board.registerUndo()
		board.cells[i].SetCenter(v)

		assert.EqualValues(t, v, board.cells[i].Center, "SetCenter failed to set expected value")
	}

	board.Reset()
	board.registerUndo()
	assert.Equal(t, state, board.history[len(board.history)-1], "initial state undo did not work")
}

func BenchmarkBoardCheckBoxes(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sudoku")

	cb := func(cells []*cell) {
		assert.Equal(b, len(cells), 0)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.checkBoxRepeat(cb)
	}
}

func TestBoardCheckBoxesExhaustive(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	var offset int
	var ic, jc *cell
	for box := 0; box < board.boxesWide*board.boxesTall; box++ {
		offset = box * board.cellsPerBox

		// ! should do half as many comparisons, but I need to test which ones
		for i := 0; i < board.cellsPerBox; i++ {
			for j := 0; j < board.cellsPerBox; j++ {
				if i == j {
					continue
				}

				ic = board.cells[offset+i]
				jc = board.cells[offset+j]
				ic.SetCenter("5")
				jc.SetCenter("5")
				assert.Error(t, board.check(), "failed check for duplicate (%d, %d) in box %d", i, j, box)

				ic.SetCenter("")
				jc.SetCenter("")
				assert.NoError(t, board.check())
			}
		}
	}
}
func BenchmarkBoardCheckCols(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sudoku")

	cb := func(cells []*cell) {
		assert.Equal(b, len(cells), 0)
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

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sudoku")

	cb := func(cells []*cell) {
		assert.Equal(b, len(cells), 0)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.checkRowRepeat(cb)
	}
}

func BenchmarkBoardCheck(b *testing.B) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(b, board.load(wikipedia), "failed loading valid classic sudoku")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		board.check()
	}
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
