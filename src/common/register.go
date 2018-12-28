package common

import "time"

type Register struct {
	Name    string
	Reg     uint64
	dataBus *uint64
	width   int
	mask    uint64
	// For rendering
	X          float64
	Y          float64
	lastUpdate time.Time
}

func (r *Register) Init(dataBus *uint64, width int, name string) {
	r.dataBus = dataBus
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}
	r.Name = name
	r.X = 0
	r.Y = 0
	r.lastUpdate = time.Now()
}

func (r *Register) Read() {
	value := r.Reg & r.mask
	*r.dataBus = value
}

func (r *Register) Write() {
	r.Reg = *r.dataBus & r.mask
	r.lastUpdate = time.Now()
}
