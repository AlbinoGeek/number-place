package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:generate go run gen.go

// TODO: These become configuration

// HighlightMistakes will become a configuration variable representing whether
// summary mistakes are checked for with every set.
var HighlightMistakes = true

var wikipedia = `3,3,3,3,53-6---98-7-195----------6-8--4--7---6-8-3-2---3--1--6-6----------419-8-28---5-79`

func main() {
	a := app.NewWithID("com.github.albinogeek.number-place")
	b, _ := start(a)
	b.load(wikipedia)
	a.Run()
}

func start(a fyne.App) (*board, fyne.Window) {
	w := a.NewWindow("Number Place")
	b := newBoard(3, 3, 3, 3)

	uiInit(b, w)

	w.Show()

	return b, w
}

func uiInit(b *board, w fyne.Window) {
	controls := make([]fyne.CanvasObject, 0)

	values := make(map[string]bool)
	// TODO: Support other digit systems (such as HEX for sandwiche or Giant)
	for i := 0; i < b.boxWidth*b.boxesWide; i++ {
		v := strconv.Itoa(1 + i)
		controls = append(controls, widget.NewButton(v, setSelected(b, v)))
		values[v] = true
	}

	w.Canvas().SetOnTypedKey(keyHandler(b, values))

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
			widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
				if err := b.check(); err != nil {
					dialog.ShowError(err, w)
					return
				}

				dialog.ShowInformation("Check Passed",
					"I don't see any mistakes right now.\nIt's up to you to complete the puzzle.", w)
			}),
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

func keyHandler(b *board, values map[string]bool) func(*fyne.KeyEvent) {
	key := func(value string) {
		if yes, ok := values[value]; !ok || !yes {
			return
		}

		setSelected(b, value)()
	}

	return func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.Key1:
			key("1")
		case fyne.Key2:
			key("2")
		case fyne.Key3:
			key("3")
		case fyne.Key4:
			key("4")
		case fyne.Key5:
			key("5")
		case fyne.Key6:
			key("6")
		case fyne.Key7:
			key("7")
		case fyne.Key8:
			key("8")
		case fyne.Key9:
			key("9")
		case fyne.KeyDelete:
			clearSelected(b)()
		case fyne.KeyBackspace:
			clearSelected(b)()
		}
	}
}

func clearSelected(b *board) func() {
	return func() {
		setSelected(b, "")()
	}
}

func setSelected(b *board, value string) func() {
	return func() {
		b.registerUndo()

		b.mu.Lock()
		for _, c := range b.cells {
			if c.selected {
				c.selected = false
				c.SetCenter(value)
			} else {
				c.SetMistake(false)
			}
		}
		b.mu.Unlock()

		b.check()
	}
}
