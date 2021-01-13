package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

type board struct {
	*fyne.Container
	cells []*cell

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
	return nil
}

func (b *board) init() {
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

	for i, c := range b.cells {
		if v := parts[4][i]; v != '-' {
			c.SetGiven(string(v))
		}
	}

	return b.check()
}
