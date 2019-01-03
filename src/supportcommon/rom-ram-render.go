package supportcommon

import (
	"common"
	"css"
	"fmt"
	"image"

	"github.com/tfriedel6/canvas"
)

// RamRomRenderer contains all the rendering code of our cpu
type RamRomRenderer struct {
	core      *RamRom
	dirty     int
	bounds    image.Rectangle
	busHeight int
	busWidth  int

	// Child Renderers
	dataBusRenderer common.BusRenderer
	busBufRenderer  common.BufferRenderer
	addrRenderer    common.RegisterRenderer
	dataRenderers   []common.RegisterRenderer
}

// InitRender Initializes the renderer
func (r *RamRomRenderer) InitRender(core *RamRom, canvas *canvas.Canvas, bounds image.Rectangle) {
	r.core = core
	// Bounds will only contain the top left corner. We will compute the size in here
	r.bounds = bounds
	r.dirty = 2

	// Initialize all the child renderers
	r.busHeight = 80
	r.busWidth = 20

	r.bounds.Max = r.bounds.Min.Add(image.Point{int(css.RegisterWidth * 2), r.busHeight + 5*int(css.RegisterHeight)})

	r.addrRenderer.InitRender(&r.core.addressReg, image.Rectangle{
		image.Point{r.bounds.Min.X, r.bounds.Min.Y},
		image.Point{r.bounds.Max.X, r.bounds.Min.Y + int(css.RegisterHeight)}})

	r.dataRenderers = make([]common.RegisterRenderer, 3)
	for i := range r.dataRenderers {
		r.dataRenderers[i].InitRender(&r.core.valueRegisters[i], image.Rectangle{
			image.Point{r.bounds.Min.X, (i+1)*int(css.RegisterHeight) + r.bounds.Min.Y},
			image.Point{r.bounds.Max.X,
				(i+2)*int(css.RegisterHeight) + r.bounds.Min.Y},
		})
		r.dataRenderers[i].ShowUpdates = false
	}

	r.busBufRenderer.InitRender(&r.core.busBuf, image.Rectangle{
		image.Point{r.bounds.Min.X, r.bounds.Max.Y - r.busHeight - int(css.RegisterHeight)},
		image.Point{r.bounds.Max.X, r.bounds.Max.Y - r.busHeight}})

	r.dataBusRenderer.InitRender(r.core.busInt,
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Max.Y - r.busHeight},
		image.Point{r.bounds.Max.X - int(css.RegisterWidth), r.bounds.Max.Y},
		r.busWidth)

}

func (r *RamRomRenderer) Bounds() image.Rectangle {
	return r.bounds
}

// Render the contents to the screen
func (r *RamRomRenderer) Render(canvas *canvas.Canvas) {

	r.dataBusRenderer.DrivingBus = &r.core.drivingBus
	r.dataBusRenderer.Render(canvas)
	r.addrRenderer.Render(canvas)
	r.busBufRenderer.Render(canvas)
	for i := range r.dataRenderers {
		r.dataRenderers[i].Render(canvas)
	}

	canvas.SetFillStyle("#000")
	canvas.FillText(fmt.Sprintf("ROM %X", r.core.chipID), float64(r.bounds.Min.X+20), float64(r.bounds.Max.Y-20))
}
