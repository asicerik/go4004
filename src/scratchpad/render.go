package scratchpad

import (
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Render the contents to the screen
func (r *Registers) Render(canvas *canvas.Canvas) {
	// Save the current state
	canvas.Save()

	w := css.RegisterWidth
	h := css.RegisterHeight

	// Render the busses first
	busHeight := 80
	busWidth := 20

	canvas.Translate(float64(canvas.Width())-w*2-css.Margin, css.Margin+float64(busWidth))

	r.renderBus.InitRender(image.Point{int(css.RegisterWidth), 0}, image.Point{int(css.RegisterWidth), busHeight}, busWidth)
	r.renderBus.Render(canvas)

	for i := 0; i < 16; i++ {
		r.regs[i].Render(canvas, css.RegisterWidth, css.RegisterHeight)
	}

	canvas.SetFillStyle("#000")
	canvas.FillText("Scratch Pad", float64(30), float64(busHeight)+8*h+30)

	canvas.Stroke()
	// Restore the state
	canvas.Restore()
}
