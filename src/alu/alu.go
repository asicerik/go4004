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
}

const FlagPosZero = uint64(0x2)
const FlagPosCarry = uint64(0x4)

type AluFlags struct {
	Zero  int // The accumulator is zero
	Carry int // The carry bit is set
}

// From the instruction decoder
const AluIntModeNone = 0
const AluIntModeAdd = 1
const AluIntModeSub = 2

const AluAdd = "+"
const AluSub = "-"
const AluNone = ""

// Init initialize the ALU
func (a *Alu) Init(dataBus *common.Bus, width int) {
	a.dataBus = dataBus
	a.width = width
	for i := 0; i < width; i++ {
		a.mask = a.mask << 1
		a.mask = a.mask | 1
	}
	a.aluCore.Init(dataBus, width)
	a.accumulator.Init(dataBus, width, "ACC=")
	a.tempRegister.Init(dataBus, width, "Temp=")
	a.flagRegister.Init(dataBus, width, "Flags=")
	a.updateFlags()
}

func (a *Alu) WriteAccumulator() {
	rlog.Debugf("Wrote Accumulator with 0x%X", a.dataBus.Read())
	a.accumulator.Write()
	a.updateFlags()
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

func (a *Alu) ReadFlags() {
	a.flagRegister.Read()
	a.tempDrivingBus = true

}

func (a *Alu) SetMode(mode int) {
	switch mode {
	case AluIntModeNone:
		a.aluCore.SetMode(AluNone)
	case AluIntModeAdd:
		a.aluCore.SetMode(AluAdd)
	case AluIntModeSub:
		a.aluCore.SetMode(AluSub)
	default:
		rlog.Warnf("** Invalid ALU mode %d", mode)
	}
}

func (a *Alu) Evaluate() {
	a.aluCore.Evaluate(a.accumulator.ReadDirect(), a.tempRegister.ReadDirect())
}

func (a *Alu) ReadEval() {
	a.aluCore.ReadOutput()
}

func (a *Alu) GetFlags() AluFlags {
	flagsRaw := a.flagRegister.ReadDirect()
	flags := AluFlags{}
	if (flagsRaw & FlagPosZero) != 0 {
		flags.Zero = 1
	}
	if (flagsRaw & FlagPosCarry) != 0 {
		flags.Carry = 1
	}
	return flags
}

func (a *Alu) updateFlags() {
	accum := a.accumulator.ReadDirect()
	flags := a.flagRegister.ReadDirect()
	if accum == 0 {
		flags |= FlagPosZero
	} else {
		flags &= ^FlagPosZero
	}
	if a.aluCore.Carry != 0 {
		flags |= FlagPosCarry
	} else {
		flags &= ^FlagPosCarry
	}
	a.flagRegister.WriteDirect(flags)
}

type aluCore struct {
	Carry     uint64 // carry flag
	outputReg common.Register
	dataBus   *common.Bus
	width     int
	mask      uint64
	carryMask uint64 // Which bit position means a carry
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
	a.carryMask = 1 << uint64(width)
	a.outputReg.Init(dataBus, width, "ALU=")
	a.mode = AluNone
	a.changed = true // force a render
}

func (a *aluCore) ReadOutput() {
	a.outputReg.Read()
}

func (a *aluCore) SetMode(mode string) {
	a.mode = mode
	rlog.Debugf("** ALU: Set mode to %s", mode)
}

func (a *aluCore) Evaluate(accIn uint64, tmpIn uint64) {
	out := accIn
	switch a.mode {
	case AluAdd:
		out = accIn + tmpIn
		if (out & a.carryMask) != 0 {
			a.Carry = 1
		} else {
			a.Carry = 0
		}
	case AluSub:
		out = accIn - tmpIn
	}
	out = out & a.mask
	a.dataBus.Reset()
	a.dataBus.Write(out)
	rlog.Debugf("** ALU: Evaluated mode %s, A=%X, T=%X, out=%X, carry=%X",
		a.mode, accIn, tmpIn, out, a.Carry)
}
