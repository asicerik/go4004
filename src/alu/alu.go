package alu

import (
	"common"

	"github.com/romana/rlog"
)

// Alu contains the processor's ALU and associated components
type Alu struct {
	aluCore         aluCore
	accumulator     common.Register
	tempRegister    common.Register
	flagRegister    common.Register
	dataBus         *common.Bus
	width           int
	mask            uint64
	accBus          common.Bus
	tempBus         common.Bus
	flagBus         common.Bus
	coreBus         common.Bus
	accumDrivingBus bool
	tempDrivingBus  bool
	flagDrivingBus  bool
	coreDrivingBus  bool
	flags           AluFlags
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
	a.aluCore.Init(dataBus, width)
	a.flags.Zero = 1
	a.accumulator.Init(dataBus, width, "ACC=")
	a.tempRegister.Init(dataBus, width, "Temp=")
	a.flagRegister.Init(dataBus, width, "Flags=")
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
	a.accumDrivingBus = true
}

func (a *Alu) WriteTemp() {
	rlog.Debugf("Wrote Temp with 0x%X", a.dataBus.Read())
	a.tempRegister.Write()
}

func (a *Alu) ReadTemp() {
	a.tempRegister.Read()
	a.tempDrivingBus = true
}

func (a *Alu) ReadAccumulatorDirect() uint64 {
	return a.accumulator.ReadDirect()
}

func (a *Alu) ReadTempDirect() uint64 {
	return a.tempRegister.ReadDirect()
}

func (a *Alu) ReadFlags() AluFlags {
	a.tempDrivingBus = true
	return a.flags
}

const AluAdd = "+"
const AluSub = "-"
const AluNone = ""

type aluCore struct {
	outputReg common.Register
	dataBus   *common.Bus
	width     int
	mask      uint64
	mode      string
	changed   bool
}

// Init initialize the ALU
func (a *aluCore) Init(dataBus *common.Bus, width int) {
	a.dataBus = dataBus
	a.width = width
	for i := 0; i < width; i++ {
		a.mask = a.mask << 1
		a.mask = a.mask | 1
	}
	a.outputReg.Init(dataBus, width, "ALU=")
	a.mode = AluNone
	a.changed = true // force a render
}
