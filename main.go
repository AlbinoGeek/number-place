package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
)

//go:generate go run gen.go

func main() {
	var (
		a = app.NewWithID("number-place.AlbinoGeek.com.github")
		w = a.NewWindow("Number Place")
	)

	uiInit(w)

	w.Show()
	a.Run()
}

func uiInit(w fyne.Window) {
	w.CenterOnScreen()
}
