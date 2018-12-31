package main

import (
	"common"
	"css"
	"fmt"
	"image"
	"math"

	"github.com/tfriedel6/canvas"
)

// IoBusRenderer renders an IO bus to the screen
type IoBusRenderer struct {
	bus    *common.Bus
	bounds image.Rectangle
	dirty  int // needs to be redrawn if non-zero
	// Child renderers
	leds []IoBitRenderer
}

// InitRender initializes the element for rendering
func (r *IoBusRenderer) InitRender(bus *common.Bus, base int, bounds image.Rectangle) {
	r.bus = bus
	r.bounds = bounds
	r.dirty = 2
	r.leds = make([]IoBitRenderer, 4)
	ledHeight := bounds.Dy() / 4
	for i := 0; i < 4; i++ {
		ledBounds := image.Rectangle{image.Point{bounds.Min.X, bounds.Min.Y + ledHeight*i},
			image.Point{bounds.Max.X, bounds.Min.Y + ledHeight*(i+1)}}
		r.leds[i].InitRender(fmt.Sprintf("IO %2d", base+i), ledBounds)
	}
}

// Render the contents to the screen
func (r *IoBusRenderer) Render(canvas *canvas.Canvas) {
	if r.bus.Updated {
		r.dirty = 2
		r.bus.Updated = false
	}
	if r.dirty == 0 {
		return
	}
	for i := range r.leds {
		data := r.bus.Read()
		bit := uint8((data >> uint64(i)) & 0x1)
		r.leds[i].Render(bit, canvas)
	}
	r.dirty--
}

// IoBitRenderer renders an IO bus bit to the screen
type IoBitRenderer struct {
	name   string
	bounds image.Rectangle
}

// InitRender initializes the element for rendering
func (r *IoBitRenderer) InitRender(name string, bounds image.Rectangle) {
	r.name = name
	r.bounds = bounds
}

// Render the contents to the screen
func (r *IoBitRenderer) Render(bit uint8, canvas *canvas.Canvas) {
	if bit == 0 {
		canvas.SetFillStyle(css.LedRedOff)
	} else {
		canvas.SetFillStyle(css.LedRedOn)
	}
	canvas.SetStrokeStyle(css.LedRedBorder)
	centerY := float64(r.bounds.Min.Y+r.bounds.Max.Y) / 2
	radius := centerY - float64(r.bounds.Min.Y) - 2
	centerX := float64(r.bounds.Max.X) - radius
	canvas.BeginPath()
	canvas.MoveTo(centerX, centerY)
	canvas.Arc(centerX, centerY, radius, 0, 2*math.Pi, false)
	canvas.ClosePath()
	canvas.Fill()
	canvas.SetFillStyle(css.TextNormal)
	canvas.FillText(fmt.Sprintf("%s", r.name), float64(r.bounds.Min.X), centerY+10)

}
