package alu

import (
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Render the contents to the screen
func (a *Alu) Render(canvas *canvas.Canvas) {
	// Save the current state
	canvas.Save()

	// Render the busses first
	busHeight := 80
	busWidth := 20

	canvas.Translate(css.Margin, css.Margin+float64(busWidth))

	a.accBus.InitRender(image.Point{int(css.RegisterWidth / 2), 0}, image.Point{int(css.RegisterWidth / 2), busHeight}, busWidth)
	a.tempBus.InitRender(image.Point{int(css.RegisterWidth/2) + 120, 0}, image.Point{int(css.RegisterWidth/2) + 120, busHeight}, busWidth)
	a.accBus.Render(canvas)
	a.tempBus.Render(canvas)

	a.accumulator.Y = float64(busHeight)
	a.tempRegister.Y = float64(busHeight)
	a.accumulator.Render(canvas, css.RegisterWidth, css.RegisterHeight)
	canvas.Translate(120, 0)
	a.tempRegister.Render(canvas, css.RegisterWidth, css.RegisterHeight)
	// Restore the state
	canvas.Restore()
}
