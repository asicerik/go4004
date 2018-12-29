package cpucore

import (
	"alu"
	"common"
	"image"
	"scratchpad"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our cpu
type Renderer struct {
	core   *Core
	dirty  int
	bounds image.Rectangle

	// Child Renderers
	internalDataBusRenderer common.BusRenderer
	aluRenderer             alu.Renderer
	scratchPadRenderer      scratchpad.Renderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(core *Core, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.core = core
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	mainBusSizePx := 40
	mainBusY := r.bounds.Min.Y + mainBusSizePx/2
	r.internalDataBusRenderer.InitRender(&r.core.internalDataBus,
		image.Point{r.bounds.Min.X, mainBusY},
		image.Point{r.bounds.Max.X, mainBusY},
		mainBusSizePx)

	aluLeftMargin := 20
	aluWidth := 400
	aluHeight := 200
	r.aluRenderer.InitRender(&r.core.alu, canvas, image.Rectangle{
		image.Point{r.bounds.Min.X + aluLeftMargin, mainBusY/2 + mainBusSizePx},
		image.Point{r.bounds.Min.X + aluWidth + aluLeftMargin, mainBusY/2 + mainBusSizePx + aluHeight}})

	spRightMargin := 20
	spWidth := 400
	spHeight := 500

	r.scratchPadRenderer.InitRender(&r.core.regs, canvas, image.Rectangle{
		image.Point{r.bounds.Max.X - spRightMargin - spWidth, mainBusY/2 + mainBusSizePx},
		image.Point{r.bounds.Max.X - spRightMargin, mainBusY/2 + mainBusSizePx + spHeight}})
}

// Render renders the current state of this element
func (r *Renderer) Render(canvas *canvas.Canvas) {
	canvas.SetFillStyle("#ccc")
	if r.dirty > 0 {
		canvas.FillRect(0, 0, float64(canvas.Width()), float64(canvas.Height()))
		r.dirty--
	}
	r.internalDataBusRenderer.Render(canvas)
	r.aluRenderer.Render(canvas)
	r.scratchPadRenderer.Render(canvas)
}
