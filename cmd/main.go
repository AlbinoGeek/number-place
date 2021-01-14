package main

import (
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

//go:generate go run gen.go

var wikipedia = `3,3,3,3,53-6---98-7-195----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`

func main() {
	var (
		a = app.NewWithID("com.github.albinogeek.number-place")
		w = a.NewWindow("Number Place")
	)

	uiInit(w)

	w.Show()
	a.Run()
}

func uiInit(w fyne.Window) {
	b := newBoard(3, 3, 3, 3)
	b.load(wikipedia)

	controls := make([]fyne.CanvasObject, 0)

	// TODO: Support other digit systems (such as HEX for sandwiche or Giant)
	for i := 0; i < b.boxWidth*b.boxesWide; i++ {
		n := 1 + i
		controls = append(controls, widget.NewButton(strconv.Itoa(n), setSelected(b, n)))
	}

	controlArea := container.NewVBox(
		fyne.NewContainerWithLayout(
			layout.NewAdaptiveGridLayout(b.boxWidth),
			controls...,
		),
		widget.NewSeparator(),
		fyne.NewContainerWithLayout(
			layout.NewAdaptiveGridLayout(3),
			widget.NewButtonWithIcon("", theme.CancelIcon(), clearSelected(b)),
			widget.NewButtonWithIcon("", theme.ContentUndoIcon(), b.undo),
		),
	)

	w.SetContent(container.NewBorder(
		nil, nil, nil,

		// Right
		controlArea,

		// Objects
		b.Container,
	))

	w.SetFixedSize(true)
	w.CenterOnScreen()
}

func clearSelected(b *board) func() {
	return func() {
		b.mu.Lock()
		for _, c := range b.cells {
			if c.selected {
				c.selected = false
				c.SetCenter("")
			}
		}
		b.mu.Unlock()
	}
}

func setSelected(b *board, n int) func() {
	return func() {
		b.registerUndo()

		b.mu.Lock()
		for _, c := range b.cells {
			if c.selected {
				c.selected = false
				c.SetCenter(strconv.Itoa(n))
			}
		}
		b.mu.Unlock()

		b.check()
	}
}
