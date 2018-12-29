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
	externalDataBusRenderer common.BusRenderer
	aluRenderer             alu.Renderer
	scratchPadRenderer      scratchpad.Renderer
	externalBufferRenderer  ExternalBusBufferRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(core *Core, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.core = core
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	mainBusSizePx := 32
	extBusY := r.bounds.Min.Y + mainBusSizePx/2
	r.externalDataBusRenderer.InitRender(&r.core.externalDataBus,
		image.Point{r.bounds.Min.X, extBusY},
		image.Point{r.bounds.Max.X, extBusY},
		mainBusSizePx)

	intBusY := extBusY + 160
	r.internalDataBusRenderer.InitRender(&r.core.internalDataBus,
		image.Point{r.bounds.Min.X, intBusY},
		image.Point{r.bounds.Max.X, intBusY},
		mainBusSizePx)

	extBufferWidth := 200
	extBufferHeight := intBusY - extBusY - mainBusSizePx
	extBufferLeft := canvas.Width()/2 - extBufferWidth/2
	extBufferTop := extBusY + mainBusSizePx/2
	r.externalBufferRenderer.InitRender(&r.core.busBuffer, canvas, image.Rectangle{
		image.Point{extBufferLeft, extBufferTop},
		image.Point{extBufferLeft + extBufferWidth, extBufferTop + extBufferHeight}})

	aluLeftMargin := 20
	aluWidth := 400
	aluHeight := 200
	r.aluRenderer.InitRender(&r.core.alu, canvas, image.Rectangle{
		image.Point{r.bounds.Min.X + aluLeftMargin, intBusY + mainBusSizePx/2},
		image.Point{r.bounds.Min.X + aluWidth + aluLeftMargin, intBusY + mainBusSizePx/2 + aluHeight}})

	spRightMargin := 20
	spWidth := 400
	spHeight := 500

	r.scratchPadRenderer.InitRender(&r.core.regs, canvas, image.Rectangle{
		image.Point{r.bounds.Max.X - spRightMargin - spWidth, intBusY + mainBusSizePx/2},
		image.Point{r.bounds.Max.X - spRightMargin, intBusY + mainBusSizePx/2 + spHeight}})
}

// Render renders the current state of this element
func (r *Renderer) Render(canvas *canvas.Canvas) {
	canvas.SetFillStyle("#ccc")
	if r.dirty > 0 {
		canvas.FillRect(0, 0, float64(canvas.Width()), float64(canvas.Height()))
		r.dirty--
	}
	r.externalDataBusRenderer.Render(canvas)
	r.internalDataBusRenderer.Render(canvas)
	r.externalBufferRenderer.Render(canvas)
	r.aluRenderer.Render(canvas)
	r.scratchPadRenderer.Render(canvas)
}

type ExternalBusBufferRenderer struct {
	buf    *ExternalBusBuffer
	dirty  int
	bounds image.Rectangle

	// Child Renderers
	internalDataBusRenderer common.BusRenderer
	externalDataBusRenderer common.BusRenderer
	bufferRenderer          common.BufferRenderer
}

// InitRender Initializes the renderer
func (r *ExternalBusBufferRenderer) InitRender(buf *ExternalBusBuffer, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.buf = buf
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	busWidthPx := 16
	busHeightPx := 40
	// bufWidth := 150
	// bufHeight := 80
	bufTop := r.bounds.Min.Y + busHeightPx
	bufHeight := r.bounds.Dy() - 2*busHeightPx
	r.externalDataBusRenderer.InitRender(r.buf.busExt,
		image.Point{r.bounds.Min.X + r.bounds.Dx()/2 - busWidthPx/2, r.bounds.Min.Y},
		image.Point{r.bounds.Min.X + r.bounds.Dx()/2 - busWidthPx/2, bufTop},
		busWidthPx)

	r.internalDataBusRenderer.InitRender(r.buf.busInt,
		image.Point{r.bounds.Min.X + r.bounds.Dx()/2 - busWidthPx/2, r.bounds.Max.Y - busHeightPx},
		image.Point{r.bounds.Min.X + r.bounds.Dx()/2 - busWidthPx/2, r.bounds.Max.Y},
		busWidthPx)

	r.bufferRenderer.InitRender(&r.buf.buf, image.Rectangle{
		image.Point{r.bounds.Min.X, bufTop},
		image.Point{r.bounds.Max.X, bufTop + bufHeight}})
}

// Render renders the current state of this element
func (r *ExternalBusBufferRenderer) Render(canvas *canvas.Canvas) {
	r.externalDataBusRenderer.Render(canvas)
	r.internalDataBusRenderer.Render(canvas)
	r.bufferRenderer.Render(canvas)
}
