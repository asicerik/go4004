package addressstack

import (
	"common"
	"css"
	"image"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our address stack
type Renderer struct {
	as     *AddressStack
	dirty  int
	bounds image.Rectangle

	// Child Renderers
	dataBusRenderer   common.BusRenderer
	registerRenderers []common.RegisterRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(as *AddressStack, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.as = as
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	busHeight := 80
	busWidth := 20

	r.dataBusRenderer.InitRender(r.as.dataBus,
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Min.Y},
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Min.Y + busHeight},
		busWidth)

	r.registerRenderers = make([]common.RegisterRenderer, len(as.stack)+1)

	for i := range r.registerRenderers {
		reg := &r.as.pc
		if i > 0 {
			reg = &r.as.stack[i-1]
		}
		r.registerRenderers[i].InitRender(reg, image.Rectangle{
			image.Point{r.bounds.Min.X, (i)*int(css.RegisterHeight) + r.bounds.Min.Y + busHeight},
			image.Point{r.bounds.Max.X,
				(i+1)*int(css.RegisterHeight) + r.bounds.Min.Y + busHeight},
		})
	}
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {

	r.dataBusRenderer.Render(canvas)

	for i := range r.registerRenderers {
		r.registerRenderers[i].Render(canvas)
	}

	canvas.SetFillStyle("#000")
	canvas.FillText("Address Stack", float64(r.bounds.Min.X+20), float64(r.bounds.Max.Y-50))
}
