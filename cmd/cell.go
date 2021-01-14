package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/driver/mobile"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var _ fyne.Widget = (*cell)(nil)
var _ desktop.Hoverable = (*cell)(nil)
var _ desktop.Mouseable = (*cell)(nil)
var _ mobile.Touchable = (*cell)(nil)

type cell struct {
	widget.BaseWidget `json:"-"`

	ID     int
	Center string
	Given  string

	hovered  bool `json:"-"`
	mistake  bool `json:"-"`
	selected bool `json:"-"`
}

func newCell(id int) *cell {
	// TODO: implement corner numbers
	// c := fyne.NewContainerWithLayout(
	// layout.NewAdaptiveGridLayout(3),
	// )

	c := &cell{ID: id}
	c.ExtendBaseWidget(c)

	return c
}

func (c *cell) CreateRenderer() fyne.WidgetRenderer {
	rend := &cellRenderer{
		cell: c,
		rect: canvas.NewRectangle(color.Transparent),
		text: canvas.NewText(c.Center, theme.TextColor()),
	}

	rend.rect.StrokeWidth = 1
	rend.rect.StrokeColor = theme.ShadowColor()
	rend.text.Alignment = fyne.TextAlignCenter
	rend.text.TextSize = 20
	rend.text.TextStyle.Monospace = true

	return rend
}

func (c *cell) MouseIn(evt *desktop.MouseEvent) {
	if c.Given == "" {
		c.hovered = true

		if evt.Button == desktop.LeftMouseButton && downCell != c {
			wasSelected = false // reset drag event
			downCell = c
			c.Select()
			return
		}

		c.Refresh()
	}
}

func (c *cell) MouseOut() {
	if c.Given == "" {
		c.hovered = false
		c.Refresh()
	}
}

// TODO: MOVE THIS, RENAME THIS
var downCell *cell
var wasSelected bool

func (c *cell) MouseMoved(*desktop.MouseEvent) {}

func (c *cell) MouseDown(*desktop.MouseEvent) {
	downCell = c
	wasSelected = c.selected

	if !c.selected {
		c.Select()
	}
}

func (c *cell) MouseUp(*desktop.MouseEvent) {
	if downCell == c && wasSelected {
		c.selected = false
		c.Refresh()
	}

	downCell = nil
	wasSelected = false
}

func (c *cell) TouchDown(*mobile.TouchEvent) {
	c.MouseDown(nil)
}

func (c *cell) TouchUp(*mobile.TouchEvent) {
	c.MouseUp(nil)
}

func (c *cell) TouchCancel(*mobile.TouchEvent) {}

// ---

func (c *cell) Select() {
	if c.Given == "" && !c.selected {
		c.selected = true
		c.Refresh()
	}
}

func (c *cell) SetGiven(n string) {
	c.mistake = false
	c.Given = n
	c.Refresh()
}

func (c *cell) SetCenter(n string) {
	c.mistake = false
	c.Center = n
	c.Refresh()
}

func (c *cell) SetMistake(b bool) {
	c.mistake = b
	c.Refresh()
}

type cellRenderer struct {
	cell *cell
	rect *canvas.Rectangle
	text *canvas.Text
}

func (r *cellRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *cellRenderer) Destroy() {}

func (r *cellRenderer) Layout(space fyne.Size) {
	r.rect.Resize(space)

	tSize := r.text.MinSize()
	r.text.Move(fyne.NewPos(0, space.Height/2-tSize.Height/2))
	r.text.Resize(fyne.NewSize(space.Width, tSize.Height))
}

func (r *cellRenderer) MinSize() fyne.Size {
	return r.text.MinSize().Max(fyne.NewSize(48, 48))
}

func (r *cellRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect, r.text}
}

func (r *cellRenderer) Refresh() {
	if r.cell.Given != "" {
		// cell is known to be correct and cannot be modified
		r.rect.FillColor = theme.ShadowColor()
		r.rect.StrokeColor = theme.HoverColor()
		r.text.Text = r.cell.Given
	} else {
		// cell is unknown and can be selected and modified
		if r.cell.hovered {
			r.rect.FillColor = theme.HoverColor()
		} else if r.cell.selected {
			r.rect.FillColor = theme.FocusColor()
		} else {
			r.rect.FillColor = color.Transparent
		}

		if r.cell.mistake {
			r.rect.StrokeColor = theme.PrimaryColorNamed(theme.ColorRed)
		} else if r.cell.selected {
			r.rect.StrokeColor = theme.FocusColor()
		} else {
			r.rect.StrokeColor = theme.ShadowColor()
		}

		r.text.Text = r.cell.Center
	}

	r.rect.Refresh()
	r.text.Refresh()
}
