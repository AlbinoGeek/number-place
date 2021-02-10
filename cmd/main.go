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

var wikipedia = `3,3,3,3,53--7----6--195----98----6-8---6---34--8-3--17---2---6-6----28----419--5----8--79`

func main() {
	a := app.NewWithID("com.github.albinogeek.number-place")
	b, _ := start(a)
	b.load(wikipedia)
	b.gameTimer.Show()
	b.gameTimer.Refresh()
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
	for i := 0; i < b.cellsPerRow; i++ {
		v := strconv.Itoa(1 + i)
		controls = append(controls, widget.NewButton(v, setSelected(b, v)))
		values[v] = true
	}

	w.Canvas().SetOnTypedKey(keyHandler(b, values))

	controlArea := container.NewPadded(container.NewVBox(
		fyne.NewContainerWithLayout(
			layout.NewGridLayout(b.boxWidth),
			controls...,
		),
		widget.NewSeparator(),
		fyne.NewContainerWithLayout(
			layout.NewGridLayout(3),
			widget.NewButtonWithIcon("", theme.CancelIcon(), clearSelected(b)),
			widget.NewButtonWithIcon("", theme.ContentUndoIcon(), b.Undo),
			widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
				if err := b.check(); err != nil {
					dialog.ShowError(err, w)
					return
				}

				dialog.ShowInformation("Check Passed",
					"I don't see any mistakes right now.\nIt's up to you to complete the puzzle.", w)
			}),
		),
		widget.NewSeparator(),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), b.Reset),
		layout.NewSpacer(),
		container.NewHBox(b.gameTimer),
	))

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
