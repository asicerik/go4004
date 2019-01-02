package scratchpad

import (
	"common"
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our cpu
type Renderer struct {
	regs      *Registers
	dirty     int
	bounds    image.Rectangle
	busHeight int
	busWidth  int

	// Child Renderers
	dataBusRenderer   common.BusRenderer
	registerRenderers []common.RegisterRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(regs *Registers, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.regs = regs
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	r.busHeight = 80
	r.busWidth = 20

	r.dataBusRenderer.InitRender(r.regs.dataBus,
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Min.Y},
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Min.Y + r.busHeight},
		r.busWidth)

	r.registerRenderers = make([]common.RegisterRenderer, len(regs.regs))

	for i := range r.registerRenderers {
		r.registerRenderers[i].InitRender(&r.regs.regs[i], image.Rectangle{
			image.Point{i%2*int(css.RegisterWidth) + r.bounds.Max.X - 2*int(css.RegisterWidth),
				(i/2)*int(css.RegisterHeight) + r.bounds.Min.Y + r.busHeight},
			image.Point{i%2*int(css.RegisterWidth) + r.bounds.Max.X - int(css.RegisterWidth),
				(i/2)*int(css.RegisterHeight) + r.bounds.Min.Y + r.busHeight + int(css.RegisterHeight)},
		})
	}
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {
	r.dataBusRenderer.DrivingBus = &r.regs.drivingBus
	r.dataBusRenderer.Render(canvas)

	for i := range r.registerRenderers {
		r.registerRenderers[i].Render(canvas)
	}

	canvas.SetFillStyle("#000")
	canvas.FillText("Scratch Pad", float64(r.bounds.Max.X-200), float64(r.bounds.Max.Y-50))
}
