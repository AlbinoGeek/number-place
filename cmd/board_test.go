package main

import (
	"testing"

	"fyne.io/fyne/test"

	"github.com/stretchr/testify/assert"
)

// tests
var (
	testGridRepeat = `3,3,3,3,55-6---98-7-195----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`
)

func TestBoardLoadCheck(t *testing.T) {
	_ = test.NewApp()

	board := newBoard(3, 3, 3, 3)

	assert.NoError(t, board.load(wikipedia), "failed loading valid classic sodoku")
	assert.Errorf(t, board.load(testGridRepeat), "loading repeats in subgrid should have failed")
}

