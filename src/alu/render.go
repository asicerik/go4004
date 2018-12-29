package alu

import (
	"common"
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our cpu
type Renderer struct {
	alu    *Alu
	dirty  int
	bounds image.Rectangle

	// Child Renderers
	accumulatorRenderer        common.RegisterRenderer
	tempRenderer               common.RegisterRenderer
	accumulatorDataBusRenderer common.BusRenderer
	tempDataBusRenderer        common.BusRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(alu *Alu, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.alu = alu
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	busHeight := 80
	busWidth := 20
	accRegX := 20
	tmpRegX := 140

	r.accumulatorDataBusRenderer.InitRender(&r.alu.accBus,
		image.Point{int(css.RegisterWidth/2) + accRegX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + accRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.tempDataBusRenderer.InitRender(&r.alu.tempBus,
		image.Point{int(css.RegisterWidth/2) + tmpRegX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + tmpRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.accumulatorRenderer.InitRender(&r.alu.accumulator, image.Rectangle{
		image.Point{accRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{accRegX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})

	r.tempRenderer.InitRender(&r.alu.tempRegister, image.Rectangle{
		image.Point{tmpRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{tmpRegX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {
	r.accumulatorDataBusRenderer.Render(canvas)
	r.tempDataBusRenderer.Render(canvas)
	r.accumulatorRenderer.Render(canvas)
	r.tempRenderer.Render(canvas)
}
