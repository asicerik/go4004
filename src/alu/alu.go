package alu

import "common"

// Alu contains the processor's ALU and associated components
type Alu struct {
	accumulator  common.Register
	tempRegister common.Register
	dataBus      *uint8
}

// Init initialize the ALU
func (a *Alu) Init(dataBus *uint8) {
	a.dataBus = dataBus
	a.accumulator.Init(dataBus)
	a.tempRegister.Init(dataBus)
}

func (a *Alu) WriteAccumulator() {
	a.accumulator.Write()
}

func (a *Alu) ReadAccumulator() {
	a.accumulator.Read()
}
