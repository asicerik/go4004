package alu

import (
	"common"

	"github.com/romana/rlog"
)

// Alu contains the processor's ALU and associated components
type Alu struct {
	accumulator  common.Register
	tempRegister common.Register
	dataBus      *common.Bus
	width        int
	mask         uint64
	accBus       common.Bus
	tempBus      common.Bus
	flags        AluFlags
}

type AluFlags struct {
	Zero  int // The accumulator is zero
	Carry int // The carry bit is set
}

// Init initialize the ALU
func (a *Alu) Init(dataBus *common.Bus, width int) {
	a.dataBus = dataBus
	a.width = width
	for i := 0; i < width; i++ {
		a.mask = a.mask << 1
		a.mask = a.mask | 1
	}
	a.accumulator.Init(dataBus, width, "ACC=")
	a.tempRegister.Init(dataBus, width, "Temp=")
}

func (a *Alu) WriteAccumulator() {
	rlog.Debugf("Wrote Accumulator with 0x%X", a.dataBus.Read())
	a.accumulator.Write()
	accum := a.accumulator.ReadDirect()
	if accum == 0 {
		a.flags.Zero = 1
	} else {
		a.flags.Zero = 0
	}
}

func (a *Alu) ReadAccumulator() {
	a.accumulator.Read()
}

func (a *Alu) WriteTemp() {
	rlog.Debugf("Wrote Temp with 0x%X", a.dataBus.Read())
	a.tempRegister.Write()
}

func (a *Alu) ReadTemp() {
	a.tempRegister.Read()
}

func (a *Alu) ReadAccumulatorDirect() uint64 {
	return a.accumulator.ReadDirect()
}

func (a *Alu) ReadTempDirect() uint64 {
	return a.tempRegister.ReadDirect()
}

func (a *Alu) ReadFlags() AluFlags {
	return a.flags
}
