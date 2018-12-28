package common

import (
	"css"
	"fmt"
	"time"

	"github.com/tfriedel6/canvas"
)

// Render the contents to the screen
func (r *Register) Render(canvas *canvas.Canvas, w float64, h float64) {
	canvas.SetStrokeStyle(css.RegisterBorder)
	canvas.SetFillStyle(css.RegisterBackground)
	canvas.FillRect(r.X, r.Y, w, h)
	canvas.StrokeRect(r.X, r.Y, w, h)
	currTime := time.Now()
	if currTime.Sub(r.lastUpdate).Seconds() > 1.0 {
		canvas.SetFillStyle(css.RegisterTextNormal)
	} else {
		canvas.SetFillStyle(css.RegisterTextUpdate)
	}
	if r.Name != "" {
		canvas.FillText(fmt.Sprintf("%s=%X", r.Name, r.Reg), r.X+10, r.Y+30)
	} else {
		canvas.FillText(fmt.Sprintf("%X", r.Reg), r.X+50, r.Y+30)
	}
}

// Render the contents to the screen
func (b *Bus) Render(canvas *canvas.Canvas) {
	canvas.SetStrokeStyle(css.BusBackground)
	canvas.SetFillStyle(css.BusBackground)
	arrowWidth := float64(b.widthPix)
	arrowHeight := float64(b.widthPix) * 1.75
	if b.startLoc.Y == b.endLoc.Y {
		RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowWidth, arrowHeight, 0)
		RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowWidth, arrowHeight, 1)
		canvas.FillRect(float64(b.startLoc.X)+arrowWidth, float64(b.startLoc.Y-b.widthPix/2),
			float64(b.endLoc.X-b.startLoc.X)-arrowWidth*2, float64(b.widthPix))
		canvas.SetFillStyle(css.RegisterTextNormal)
		if b.Name != "" {
			canvas.FillText(fmt.Sprintf("%s=%X", b.Name, b.Data), 20+arrowWidth, 20)
		} else {
			canvas.FillText(fmt.Sprintf("%X", b.Data), 20+arrowWidth, 20)
		}
	} else if b.startLoc.X == b.endLoc.X {
		RenderArrowHead(canvas, float64(b.startLoc.X), float64(b.startLoc.Y), arrowHeight, arrowWidth, 2)
		RenderArrowHead(canvas, float64(b.endLoc.X), float64(b.endLoc.Y), arrowHeight, arrowWidth, 3)
		canvas.FillRect(float64(b.startLoc.X-b.widthPix/2), float64(b.startLoc.Y)+arrowWidth,
			float64(b.widthPix), float64(b.endLoc.Y-b.startLoc.Y)-arrowWidth*2)
	}

}

func RenderArrowHead(canvas *canvas.Canvas, x float64, y float64, w float64, h float64, dir int) {
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
	canvas.Fill()
}
