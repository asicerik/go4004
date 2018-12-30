package rom4001

import (
	"common"
	"css"
	"fmt"
	"image"

	"github.com/tfriedel6/canvas"
)

// Renderer contains all the rendering code of our cpu
type Renderer struct {
	rom       *Rom4001
	dirty     int
	bounds    image.Rectangle
	busHeight int
	busWidth  int

	// Child Renderers
	dataBusRenderer common.BusRenderer
	ioBufRenderer   common.BufferRenderer
	addrRenderer    common.RegisterRenderer
	dataRenderers   []common.RegisterRenderer
}

// InitRender Initializes the renderer
func (r *Renderer) InitRender(rom *Rom4001, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.rom = rom
	// Bounds will only contain the top left corner. We will compute the size in here
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	r.busHeight = 80
	r.busWidth = 20

	r.bounds.Max = r.bounds.Min.Add(image.Point{int(css.RegisterWidth * 2), r.busHeight + 5*int(css.RegisterHeight)})

	r.addrRenderer.InitRender(&r.rom.addressReg, image.Rectangle{
		image.Point{r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{r.bounds.Max.X, r.bounds.Min.Y + int(css.RegisterHeight)}})

	r.dataRenderers = make([]common.RegisterRenderer, 3)
	for i := range r.dataRenderers {
		r.dataRenderers[i].InitRender(&r.rom.valueRegisters[i], image.Rectangle{
			image.Point{r.bounds.Min.X, (i+1)*int(css.RegisterHeight) + r.bounds.Min.Y},
			image.Point{r.bounds.Max.X,
				(i+2)*int(css.RegisterHeight) + r.bounds.Min.Y},
		})
		r.dataRenderers[i].ShowUpdates = false
	}

	r.ioBufRenderer.InitRender(&r.rom.ioBuf, image.Rectangle{
		image.Point{r.bounds.Min.X, r.bounds.Max.Y - r.busHeight - int(css.RegisterHeight)},
		image.Point{r.bounds.Max.X, r.bounds.Max.Y - r.busHeight}})

	r.dataBusRenderer.InitRender(&r.rom.busInt,
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Max.Y - r.busHeight},
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Max.Y},
		r.busWidth)

}

func (r *Renderer) Bounds() image.Rectangle {
	return r.bounds
}

// Render the contents to the screen
func (r *Renderer) Render(canvas *canvas.Canvas) {

	r.dataBusRenderer.Render(canvas)
	r.addrRenderer.Render(canvas)
	r.ioBufRenderer.Render(canvas)
	for i := range r.dataRenderers {
		r.dataRenderers[i].Render(canvas)
	}

	canvas.SetFillStyle("#000")
	canvas.FillText(fmt.Sprintf("ROM %X", r.rom.chipID), float64(r.bounds.Min.X+20), float64(r.bounds.Max.Y-20))
}
