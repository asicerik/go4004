package cpucore

import (
	"addressstack"
	"alu"
	"common"
	"css"
	"fmt"
	"image"
	"instruction"
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
	asRenderer              addressstack.Renderer
	instRenderer            instruction.Renderer
	instDirty               int // non-zero = render
	instX                   int
	instY                   int
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(core *Core, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.core = core
	r.bounds = bounds
	r.dirty = 2
	r.instDirty = 2

	// Initialize all the child renderers
	mainBusSizePx := 32
	extBusY := r.bounds.Min.Y + mainBusSizePx/2
	r.externalDataBusRenderer.InitRender(&r.core.ExternalDataBus,
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
	aluWidth := 470
	aluHeight := 400
	aluRight := r.bounds.Min.X + aluWidth + aluLeftMargin
	r.aluRenderer.InitRender(&r.core.alu, canvas, image.Rectangle{
		image.Point{r.bounds.Min.X + aluLeftMargin, intBusY + mainBusSizePx/2},
		image.Point{aluRight, intBusY + mainBusSizePx/2 + aluHeight}})

	instLeftMargin := 20
	instLeft := aluRight + instLeftMargin
	instWidth := int(css.RegisterWidth)
	instHeight := 250
	instRight := instLeft + instWidth
	r.instRenderer.InitRender(&r.core.inst, canvas, image.Rectangle{
		image.Point{instLeft, intBusY + mainBusSizePx/2},
		image.Point{instRight, intBusY + mainBusSizePx/2 + instHeight}})

	r.instX = instLeft + 33
	r.instY = intBusY + instHeight - 40

	asLeftMargin := 40
	asLeft := instRight + asLeftMargin
	asWidth := int(2 * css.RegisterWidth)
	asHeight := 320
	r.asRenderer.InitRender(&r.core.as, canvas, image.Rectangle{
		image.Point{asLeft, intBusY + mainBusSizePx/2},
		image.Point{asLeft + asLeftMargin + asWidth, intBusY + mainBusSizePx/2 + asHeight}})

	spRightMargin := 20
	spWidth := 400
	spHeight := 500

	r.scratchPadRenderer.InitRender(&r.core.regs, canvas, image.Rectangle{
		image.Point{r.bounds.Max.X - spRightMargin - spWidth, intBusY + mainBusSizePx/2},
		image.Point{r.bounds.Max.X - spRightMargin, intBusY + mainBusSizePx/2 + spHeight}})
}

// Render renders the current state of this element
func (r *Renderer) Render(canvas *canvas.Canvas) {
	canvas.SetFillStyle(css.Background)
	if r.dirty > 0 {
		canvas.FillRect(0, 0, float64(canvas.Width()), float64(canvas.Height()))
		r.dirty--
	}
	r.externalDataBusRenderer.Render(canvas)
	r.internalDataBusRenderer.Render(canvas)
	r.externalBufferRenderer.Render(canvas)
	r.aluRenderer.Render(canvas)
	r.instRenderer.Render(canvas)
	r.asRenderer.Render(canvas)
	r.scratchPadRenderer.Render(canvas)
	if r.core.Decoder.InstChanged {
		r.instDirty = 2
	}
	if r.instDirty > 0 {
		canvas.SetFillStyle(css.Background)
		canvas.FillRect(float64(r.instX), float64(r.instY)-20, 100, 30)
		canvas.SetFillStyle(css.TextNormal)
		canvas.FillText(r.core.Decoder.DecodedInstruction, float64(r.instX), float64(r.instY))
		r.instDirty--
	}
	// Always render the clock count
	canvas.SetFillStyle(css.Background)
	canvas.FillRect(float64(r.instX), float64(r.instY), 100, 30)
	canvas.SetFillStyle(css.TextNormal)
	canvas.FillText(fmt.Sprintf("CLK=%d", r.core.GetClockCount()), float64(r.instX), float64(r.instY+20))
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
