package alu

import (
	"common"
)

// Alu contains the processor's ALU and associated components
type Alu struct {
	accumulator  common.Register
	tempRegister common.Register
	dataBus      *uint64
	width        int
	mask         uint64
	accBus       common.Bus
	tempBus      common.Bus
}

// Init initialize the ALU
func (a *Alu) Init(dataBus *uint64, width int) {
	a.dataBus = dataBus
	a.width = width
	for i := 0; i < width; i++ {
		a.mask = a.mask << 1
		a.mask = a.mask | 1
	}
	a.accumulator.Init(dataBus, width, "ACC")
	a.tempRegister.Init(dataBus, width, "Temp")
}

func (a *Alu) WriteAccumulator() {
	a.accumulator.Write()
}

func (a *Alu) ReadAccumulator() {
	a.accumulator.Read()
}
