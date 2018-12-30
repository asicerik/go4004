package instruction

import (
	"common"
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our cpu
type Renderer struct {
	instReg *Instruction
	dirty   int
	bounds  image.Rectangle

	// Child Renderers
	ioRenderer                 common.RegisterRenderer
	instructionRenderer        common.RegisterRenderer
	instructionDataBusRenderer common.BusRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(instReg *Instruction, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.instReg = instReg
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	busHeight := 80
	busWidth := 20
	regX := 20

	r.instructionDataBusRenderer.InitRender(r.instReg.dataBus,
		image.Point{int(css.RegisterWidth/2) + regX + r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{int(css.RegisterWidth/2) + regX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		busWidth)

	r.ioRenderer.InitRender(&r.instReg.busReg, image.Rectangle{
		image.Point{regX + r.bounds.Min.X, r.bounds.Min.Y + busHeight},
		image.Point{regX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + int(css.RegisterHeight)}})

	r.instructionRenderer.InitRender(&r.instReg.instReg, image.Rectangle{
		image.Point{regX + r.bounds.Min.X, r.bounds.Min.Y + busHeight + int(css.RegisterHeight)},
		image.Point{regX + r.bounds.Min.X + int(css.RegisterWidth), r.bounds.Min.Y + busHeight + 2*int(css.RegisterHeight)}})
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {
	r.instructionDataBusRenderer.DrivingBus = &r.instReg.drivingBus
	r.instructionDataBusRenderer.Render(canvas)
	r.ioRenderer.Render(canvas)
	r.instructionRenderer.Render(canvas)
}
