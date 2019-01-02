package alu

import (
	"common"

	"github.com/romana/rlog"
)

// Alu contains the processor's ALU and associated components
type Alu struct {
	aluCore         aluCore
	currentRamBank  uint64
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

// Accumulator instructions
const CLB = 0x0 // Clear accumulator and carry
const CLC = 0x1 // Clear carry
const IAC = 0x2 // Increment accumulator
const CMC = 0x3 // Complement carry
const CMA = 0x4 // Complement accumulator
const RAL = 0x5 // Rotate left (accumulator and carry)
const RAR = 0x6 // Rotate right (accumulator and carry)
const TCC = 0x7 // Transmit carry to accumulator and clear carry
const DAC = 0x8 // Decrement accumulator
const TCS = 0x9 // Transmit carry subtract and clear carry
const STC = 0xA // Set carry
const DAA = 0xB // Decimal adjust
const KBP = 0xC // Keyboard process
const DCL = 0xD // Designate command line

var instStrings = []string{"CLB", "CLC", "IAC", "CMC", "CMA", "RAL", "RAR", "TCC", "DAC", "TCS", "STC", "DAA", "KBP", "DCL"}

func accInstToString(inst uint64) string {
	return instStrings[inst]

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
	a.coreDrivingBus = true
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

func (a *Alu) GetCurrentRamBank() uint64 {
	return a.currentRamBank
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

func (a *Alu) ExectuteAccInst(inst uint64) {
	accumPre := a.accumulator.ReadDirect()
	carryPre := a.GetFlags().Carry
	// All this stuff is NOT cycle accurate. Who knows how this works
	// in the read CPU. Probably not this way though :)
	switch inst {
	case CLB:
		a.accumulator.WriteDirect(0)
		fallthrough
	case CLC:
		a.aluCore.SetCarry(0)
	case IAC:
		a.aluCore.SetMode(AluAdd)
		a.aluCore.Evaluate(a.accumulator.ReadDirect(), 1)
		a.accumulator.WriteDirect(a.aluCore.ReadOutputDirect())
	case CMC:
		a.aluCore.ComplimentCarry()
	case CMA:
		a.accumulator.WriteDirect((^(a.accumulator.ReadDirect())) & a.mask)
	case RAL:
		flags := a.GetFlags()
		accum := a.accumulator.ReadDirect()
		accum = accum << 1
		// The high bit becomes the carry bit
		if (accum & a.aluCore.carryMask) != 0 {
			a.aluCore.SetCarry(1)
		} else {
			a.aluCore.SetCarry(0)
		}
		// The low bit is the previous carry
		if flags.Carry != 0 {
			accum |= 1
		}
		a.accumulator.WriteDirect(accum)
	case RAR:
		flags := a.GetFlags()
		accum := a.accumulator.ReadDirect()
		lsb := accum & 0x1
		accum = accum >> 1
		// Set the carry to the lsb before the shift
		a.aluCore.SetCarry(lsb)
		// The high bit is the previous carry
		if flags.Carry != 0 {
			accum |= 0x8
		}
		a.accumulator.WriteDirect(accum)
	case TCC:
		flags := a.GetFlags()
		if flags.Carry != 0 {
			a.accumulator.WriteDirect(1)
		} else {
			a.accumulator.WriteDirect(0)
		}
		a.aluCore.SetCarry(0)
	case DAC:
		a.aluCore.SetMode(AluSub)
		a.aluCore.Evaluate(a.accumulator.ReadDirect(), 1)
		a.accumulator.WriteDirect(a.aluCore.ReadOutputDirect())
	case STC:
		a.aluCore.SetCarry(1)
	case DAA:
		accum := a.accumulator.ReadDirect()
		flags := a.GetFlags()
		if accum > 9 || flags.Carry != 0 {
			accum += 6
			// This command does not reset the carry, only sets it
			if accum&a.aluCore.carryMask != 0 {
				a.aluCore.SetCarry(1)
				accum = accum & a.mask
			}
			a.accumulator.WriteDirect(accum)
		}
	case KBP:
		accum := a.accumulator.ReadDirect()
		if accum < 3 {
			// Do nothing
		} else if accum == 4 {
			a.accumulator.WriteDirect(3)
		} else if accum == 8 {
			a.accumulator.WriteDirect(4)
		} else {
			a.accumulator.WriteDirect(0xf)
		}
	case DCL:
		// This command does not actually modify the accumulator
		a.currentRamBank = a.accumulator.ReadDirect() & 0x7
	}
	a.updateFlags()
	accumPost := a.accumulator.ReadDirect()
	carryPost := a.GetFlags().Carry
	cmdString := accInstToString(inst)

	rlog.Debugf("Accumulator CMD %s: accum pre=%X, carryPre=%X, accum post=%X, carryPost=%X",
		cmdString, accumPre, carryPre, accumPost, carryPost)
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

func (a *aluCore) ReadOutputDirect() uint64 {
	return a.outputReg.ReadDirect()
}

func (a *aluCore) SetCarry(val uint64) {
	a.Carry = val
}

func (a *aluCore) ComplimentCarry() {
	if a.Carry == 0 {
		a.Carry = 1
	} else {
		a.Carry = 0
	}
}

func (a *aluCore) SetMode(mode string) {
	a.mode = mode
	a.changed = true
	rlog.Debugf("** ALU: Set mode to %s", mode)
}

func (a *aluCore) Evaluate(accIn uint64, tmpIn uint64) {
	out := accIn
	prevCarry := a.Carry
	switch a.mode {
	case AluAdd:
		out = accIn + tmpIn
		if (out & a.carryMask) != 0 {
			a.Carry = 1
		} else {
			a.Carry = 0
		}
	case AluSub:
		// We set the carry bit to indicate NO borrow
		if tmpIn > accIn {
			a.Carry = 0
		} else {
			a.Carry = 1
		}
		out = accIn - tmpIn
		if prevCarry != 0 {
			out++
		}
	}
	out = out & a.mask
	a.outputReg.WriteDirect(out)
	rlog.Debugf("** ALU: Evaluated mode %s, A=%X, T=%X, carryIn=%X, out=%X, carry=%X",
		a.mode, accIn, tmpIn, prevCarry, out, a.Carry)
}
