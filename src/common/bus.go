package common

import "image"

type Bus struct {
	Name     string
	Data     uint64
	BusWidth int

	// These are for rendering
	startLoc image.Point
	endLoc   image.Point
	widthPix int
}

func (b *Bus) Init(name string) {
	b.Name = name
}

func (b *Bus) InitRender(startLoc image.Point, endLoc image.Point, widthPix int) {
	b.startLoc = startLoc
	b.endLoc = endLoc
	b.widthPix = widthPix
}
