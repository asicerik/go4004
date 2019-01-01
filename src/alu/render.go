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
	flagRenderer               common.RegisterRenderer
	accumulatorDataBusRenderer common.BusRenderer
	tempDataBusRenderer        common.BusRenderer
	flagDataBusRenderer        common.BusRenderer
	aluCoreRenderer            CoreRenderer
	accumCoreDataBusRenderer   common.BusRenderer
	tempCoreDataBusRenderer    common.BusRenderer
	aluCoreDataBusRenderer     common.BusRenderer
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
	tmpRegX := 150
	flagRegX := 280
	aluCoreX := tmpRegX + int(css.RegisterWidth)
	aluCoreY := 150
	aluCoreW := 125
	aluCoreH := 150

	r.accumulatorDataBusRenderer.InitRender(&r.alu.accBus,
		image.Point{int(css.RegisterWidth/2) + accRegX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + accRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.tempDataBusRenderer.InitRender(&r.alu.tempBus,
		image.Point{int(css.RegisterWidth/2) + tmpRegX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + tmpRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.flagDataBusRenderer.InitRender(&r.alu.flagBus,
		image.Point{int(css.RegisterWidth/2) + flagRegX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + flagRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.accumulatorRenderer.InitRender(&r.alu.accumulator, image.Rectangle{
		image.Point{accRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{accRegX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})

	r.tempRenderer.InitRender(&r.alu.tempRegister, image.Rectangle{
		image.Point{tmpRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{tmpRegX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})

	r.flagRenderer.InitRender(&r.alu.flagRegister, image.Rectangle{
		image.Point{flagRegX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{flagRegX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})

	r.aluCoreRenderer.InitRender(&r.alu.aluCore, canvas, image.Rectangle{
		image.Point{aluCoreX + r.bounds.Min.X, r.bounds.Min.Y + aluCoreY},
		image.Point{aluCoreX + r.bounds.Min.X + aluCoreW, r.bounds.Min.Y + aluCoreY + aluCoreH}})

	coreBusTop := r.bounds.Min.Y + busHeight + int(css.RegisterHeight)
	r.accumCoreDataBusRenderer.InitRender(&r.alu.accBus,
		image.Point{int(css.RegisterWidth/2) + accRegX + r.bounds.Min.X, coreBusTop},
		image.Point{aluCoreX + r.bounds.Min.X, r.bounds.Min.Y + aluCoreY + int(float64(aluCoreH)*0.80)},
		busWidth)
	r.accumCoreDataBusRenderer.NoStartArrow = true

	r.tempCoreDataBusRenderer.InitRender(&r.alu.tempBus,
		image.Point{int(css.RegisterWidth/2) + tmpRegX + r.bounds.Min.X, coreBusTop},
		image.Point{aluCoreX + r.bounds.Min.X, r.bounds.Min.Y + aluCoreY + int(float64(aluCoreH)*0.20)},
		busWidth)
	r.tempCoreDataBusRenderer.NoStartArrow = true

	r.aluCoreDataBusRenderer.InitRender(&r.alu.tempBus,
		image.Point{aluCoreX + r.bounds.Min.X + aluCoreW, r.bounds.Min.Y + aluCoreY + aluCoreH/2},
		image.Point{r.bounds.Max.X, r.bounds.Min.Y},
		busWidth)
	r.aluCoreDataBusRenderer.NoStartArrow = true
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {
	r.accumulatorDataBusRenderer.DrivingBus = &r.alu.accumDrivingBus
	r.tempDataBusRenderer.DrivingBus = &r.alu.tempDrivingBus
	r.flagDataBusRenderer.DrivingBus = &r.alu.flagDrivingBus
	r.aluCoreDataBusRenderer.DrivingBus = &r.alu.coreDrivingBus
	r.accumulatorDataBusRenderer.Render(canvas)
	r.tempDataBusRenderer.Render(canvas)
	r.flagDataBusRenderer.Render(canvas)
	r.accumulatorRenderer.Render(canvas)
	r.tempRenderer.Render(canvas)
	r.flagRenderer.Render(canvas)
	r.aluCoreRenderer.Render(canvas)
	r.accumCoreDataBusRenderer.Render(canvas)
	r.tempCoreDataBusRenderer.Render(canvas)
	r.aluCoreDataBusRenderer.Render(canvas)
}

// CoreRenderer contains all the rendering code of our cpu
type CoreRenderer struct {
	aluCore *aluCore
	bounds  image.Rectangle
	dirty   int // if non-zero, render
}

// InitRender Initializes the renderer
func (r *CoreRenderer) InitRender(aluCore *aluCore, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.aluCore = aluCore
	r.bounds = bounds
}

// Render the contents to the screen
func (r *CoreRenderer) Render(canvas *canvas.Canvas) {
	if r.aluCore.changed {
		r.dirty = 2
		r.aluCore.changed = false
	}
	if r.dirty == 0 {
		return
	}
	canvas.SetFillStyle(css.AluCoreFill)
	canvas.SetStrokeStyle(css.RegisterBorder)
	left := float64(r.bounds.Min.X)
	top := float64(r.bounds.Min.Y)
	width := float64(r.bounds.Dx())
	height := float64(r.bounds.Dy())
	canvas.BeginPath()
	canvas.MoveTo(left, top)
	canvas.LineTo(left, top+height*0.35)
	canvas.LineTo(left+width*0.4, top+height*0.5)
	canvas.LineTo(left, top+height*0.65)
	canvas.LineTo(left, top+height)
	canvas.LineTo(left+width, top+height*0.7)
	canvas.LineTo(left+width, top+height*0.3)
	canvas.ClosePath()
	canvas.Fill()

	canvas.SetFillStyle(css.RegisterTextNormal)
	canvas.FillText(r.aluCore.mode, left+width*0.65, top+height*0.5+5)
	r.dirty--
}
