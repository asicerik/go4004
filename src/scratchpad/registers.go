package scratchpad

import (
	"common"
	"css"
)

type Registers struct {
	// interfaces.BusDriver
	// interfaces.ClockedElement
	regs      []common.Register
	index     int
	dataBus   *uint64
	width     int
	mask      uint64
	renderBus common.Bus
}

func (r *Registers) Init(dataBus *uint64, width int, depth int) {
	r.regs = make([]common.Register, depth)
	for i := 0; i < depth; i++ {
		r.regs[i].Init(dataBus, width, "")
	}
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}

	r.index = 0
	r.dataBus = dataBus

	for i := 0; i < depth; i++ {
		r.regs[i].X = float64(i%2) * css.RegisterWidth
		r.regs[i].Y = float64(i/2)*css.RegisterHeight + 80
	}
}

func (r *Registers) Read() {
	r.regs[r.index].Read()
}

func (r *Registers) Select(index int) {
	r.index = index
}

func (r *Registers) Write() {
	r.regs[r.index].Write()
}
