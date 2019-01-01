package common

import (
	"css"
	"fmt"
	"image"

	"github.com/tfriedel6/canvas"
)

// RegisterRenderer renders a Register to the screen
type RegisterRenderer struct {
	reg         *Register
	bounds      image.Rectangle
	ShowUpdates bool
	dirty       int // needs to be redrawn if non-zero
}

// InitRender initializes the element for rendering
func (r *RegisterRenderer) InitRender(reg *Register, bounds image.Rectangle) {
	r.reg = reg
	r.bounds = bounds
	r.dirty = 2
	r.ShowUpdates = true
}

// Render the contents to the screen
func (r *RegisterRenderer) Render(canvas *canvas.Canvas) {
	if r.reg.changed {
		r.dirty = 4
		r.reg.changed = false
	}
	if r.dirty == 0 {
		return
	}
	if r.reg.Selected {
		canvas.SetFillStyle(css.RegisterBackgroundSelected)
	} else {
		canvas.SetFillStyle(css.RegisterBackground)
	}
	canvas.SetStrokeStyle(css.RegisterBorder)
	canvas.FillRect(float64(r.bounds.Min.X), float64(r.bounds.Min.Y),
		float64(r.bounds.Dx()), float64(r.bounds.Dy()))
	canvas.StrokeRect(float64(r.bounds.Min.X), float64(r.bounds.Min.Y),
		float64(r.bounds.Dx()), float64(r.bounds.Dy()))

	if r.dirty < 3 || !r.ShowUpdates {
		canvas.SetFillStyle(css.RegisterTextNormal)
	} else {
		canvas.SetFillStyle(css.RegisterTextUpdate)
	}
	if r.reg.Name != "" {
		canvas.FillText(fmt.Sprintf("%s%X", r.reg.Name, r.reg.Reg), float64(r.bounds.Min.X+10), float64(r.bounds.Min.Y+30))
	} else {
		canvas.FillText(fmt.Sprintf("%X", r.reg.Reg), float64(r.bounds.Min.X+50), float64(r.bounds.Min.Y+30))
	}
	r.dirty--
}

// BufferRenderer renders a Buffer to the screen
type BufferRenderer struct {
	buf    *Buffer
	bounds image.Rectangle
	dirty  int // needs to be redrawn if non-zero
}

// InitRender initializes the element for rendering
func (r *BufferRenderer) InitRender(buf *Buffer, bounds image.Rectangle) {
	r.buf = buf
	r.bounds = bounds
	r.dirty = 2
}

// Render the contents to the screen
func (r *BufferRenderer) Render(canvas *canvas.Canvas) {
	if r.buf.changed {
		r.dirty = 2
		r.buf.changed = false
	}
	if r.dirty == 0 {
		return
	}
	canvas.SetStrokeStyle(css.RegisterBorder)
	canvas.SetFillStyle(css.RegisterBackground)
	canvas.FillRect(float64(r.bounds.Min.X), float64(r.bounds.Min.Y),
		float64(r.bounds.Dx()), float64(r.bounds.Dy()))
	canvas.StrokeRect(float64(r.bounds.Min.X), float64(r.bounds.Min.Y),
		float64(r.bounds.Dx()), float64(r.bounds.Dy()))

	canvas.SetFillStyle(css.RegisterTextNormal)
	canvas.FillText(fmt.Sprintf("%s", r.buf.Name), float64(r.bounds.Min.X+10), float64(r.bounds.Min.Y+30))

	canvas.SetFillStyle(css.BufferDirArrow)
	arrowWidth := 16.0
	arrowHeight := 24.0
	// TODO: fix these magic #s here
	if r.buf.Dir == DirAtoB {
		RenderArrowHead(canvas, float64(r.bounds.Max.X)-arrowWidth-5, float64(r.bounds.Max.Y)-15, arrowHeight, arrowWidth, 3)
	} else if r.buf.Dir == DirBtoA {
		RenderArrowHead(canvas, float64(r.bounds.Max.X)-arrowWidth-5, float64(r.bounds.Max.Y)-arrowHeight-10, arrowHeight, arrowWidth, 2)
	}
	r.dirty--
}

// BusRenderer renders a Bus to the screen
type BusRenderer struct {
	bus          *Bus
	startLoc     image.Point
	endLoc       image.Point
	widthPix     int
	dirty        int  // needs to be redrawn if non-zero
	NoStartArrow bool // Don't draw the start arrowhead
	NoEndArrow   bool // Don't draw the end arrowhead
	DrivingBus   *bool
}

// InitRender initializes the element for rendering
func (b *BusRenderer) InitRender(bus *Bus, startLoc image.Point, endLoc image.Point, widthPix int) {
	b.bus = bus
	b.startLoc = startLoc
	b.endLoc = endLoc
	b.widthPix = widthPix
	b.dirty = 2
}

// Render the contents to the screen
func (b *BusRenderer) Render(canvas *canvas.Canvas) {
	busCollision := false
	busDriven := false
	if b.DrivingBus != nil {
		busDriven = *b.DrivingBus
	}
	if b.bus.writes > 0 || busDriven {
		b.dirty = 4
		if b.bus.writes > 1 {
			busCollision = true
		}
		b.bus.writes = 0
		if b.DrivingBus != nil && *b.DrivingBus {
			*b.DrivingBus = false
		}
	}

	if b.dirty == 0 {
		return
	}
	if busCollision {
		canvas.SetStrokeStyle(css.BusCollision)
		canvas.SetFillStyle(css.BusCollision)
	} else if busDriven {
		canvas.SetStrokeStyle(css.BusBackgroundWritten)
		canvas.SetFillStyle(css.BusBackgroundWritten)
	} else {
		canvas.SetStrokeStyle(css.BusBackground)
		canvas.SetFillStyle(css.BusBackground)
	}
	arrowWidth := float64(b.widthPix)
	arrowHeight := float64(b.widthPix) * 1.75
	if b.startLoc.Y == b.endLoc.Y {
		if !b.NoStartArrow {
			RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowWidth, arrowHeight, 0)
		}
		if !b.NoEndArrow {
			RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowWidth, arrowHeight, 1)
		}
		canvas.FillRect(float64(b.startLoc.X)+arrowWidth, float64(b.startLoc.Y-b.widthPix/2),
			float64(b.endLoc.X-b.startLoc.X)-arrowWidth*2, float64(b.widthPix))
		canvas.SetFillStyle(css.RegisterTextNormal)
		if b.bus.Name != "" {
			canvas.FillText(fmt.Sprintf("%s=%X", b.bus.Name, b.bus.data),
				float64(b.startLoc.X)+20+arrowWidth, float64(b.startLoc.Y)+5)
		} else {
			canvas.FillText(fmt.Sprintf("%X", b.bus.data),
				float64(b.startLoc.X)+20+arrowWidth, float64(b.startLoc.Y)+5)
		}
	} else if b.startLoc.X == b.endLoc.X {
		if !b.NoStartArrow {
			RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowHeight, arrowWidth, 2)
		}
		if !b.NoEndArrow {
			RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowHeight, arrowWidth, 3)
		}
		canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.startLoc.Y)+arrowWidth,
			float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y)-arrowWidth*2)
	} else if b.startLoc.X < b.endLoc.X {
		if b.startLoc.Y < b.endLoc.Y {
			if !b.NoStartArrow {
				RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowHeight, arrowWidth, 2)
				canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.startLoc.Y)+arrowWidth,
					float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y)-arrowWidth)
			} else {
				canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.startLoc.Y),
					float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y))
			}
			if !b.NoEndArrow {
				RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowWidth, arrowHeight, 1)
				canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.endLoc.Y-b.widthPix/2),
					float64(b.endLoc.X-b.startLoc.X), float64(b.widthPix))
			} else {
				canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.endLoc.Y-b.widthPix/2),
					float64(b.endLoc.X-b.startLoc.X+b.widthPix/2), float64(b.widthPix))
			}
		} else {
			if !b.NoStartArrow {
				RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowWidth, arrowHeight, 0)
				canvas.FillRect(float64(b.startLoc.X)+arrowWidth, float64(b.startLoc.Y-b.widthPix/2),
					float64(b.endLoc.X-b.startLoc.X)-arrowWidth, float64(b.widthPix))
			} else {
				canvas.FillRect(float64(b.startLoc.X), float64(b.startLoc.Y-b.widthPix/2),
					float64(b.endLoc.X-b.startLoc.X), float64(b.widthPix))
			}
			if !b.NoEndArrow {
				RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowHeight, arrowWidth, 2)
				canvas.FillRect(float64(b.endLoc.X-b.widthPix/2), float64(b.startLoc.Y+b.widthPix/2),
					float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y))
			} else {
				canvas.FillRect(float64(b.endLoc.X-b.widthPix/2), float64(b.startLoc.Y+b.widthPix/2),
					float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y-b.widthPix/2))
			}
		}
	}
	b.dirty--
}

func RenderArrowHead(canvas *canvas.Canvas, x float64, y float64, w float64, h float64, dir int) {
	canvas.BeginPath()
	canvas.MoveTo(x, y)
	switch dir {
	case 0:
		canvas.LineTo(x+w, y-h/2)
		canvas.LineTo(x+w, y+h/2)
		canvas.LineTo(x, y)
	case 1:
		canvas.LineTo(x-w, y-h/2)
		canvas.LineTo(x-w, y+h/2)
		canvas.LineTo(x, y)
	case 2:
		canvas.LineTo(x+w/2, y+h)
		canvas.LineTo(x-w/2, y+h)
		canvas.LineTo(x, y)
	case 3:
		canvas.LineTo(x+w/2, y-h)
		canvas.LineTo(x-w/2, y-h)
		canvas.LineTo(x, y)
	}
	canvas.ClosePath()
	canvas.Fill()
}
