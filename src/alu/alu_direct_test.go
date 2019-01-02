package alu

import (
	"common"
	"os"
	"testing"

	"github.com/romana/rlog"
)

const enableLog = true

func SetupLogger() {
	if enableLog {
		// Programmatically change an rlog setting from within the program
		os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
		os.Setenv("RLOG_TRACE_LEVEL", "0")
		os.Setenv("RLOG_LOG_FILE", "alu_test.log")
		rlog.UpdateEnv()
	}
}

func writeBus(val uint64, bus *common.Bus, t *testing.T) {
	bus.Reset()
	bus.Write(val)
	bus.Reset()
}

func writeAccumulator(val uint64, alu *Alu, bus *common.Bus, t *testing.T) {
	writeBus(val, bus, t)
	alu.WriteAccumulator()
}

func writeTemp(val uint64, alu *Alu, bus *common.Bus, t *testing.T) {
	writeBus(val, bus, t)
	alu.WriteTemp()
}

func verifyRegister(val uint64, bus *common.Bus, t *testing.T) {
	bus.Reset()
	data := bus.Read()
	if data != val {
		t.Errorf("ALU register mismatch. Exp %X, got %X", val, data)
	}
	bus.Reset()
}

func verifyFlags(val uint64, alu *Alu, bus *common.Bus, t *testing.T) {
	alu.ReadFlags()
	bus.Reset()
	data := bus.Read()
	if data != val {
		if (data & FlagPosCarry) != (val & FlagPosCarry) {
			t.Errorf("ALU carry flag mismatch. Exp %X, got %X", val&FlagPosCarry, data&FlagPosCarry)
		}
		if (data & FlagPosZero) != (val & FlagPosZero) {
			t.Errorf("ALU zero flag mismatch. Exp %X, got %X", val&FlagPosZero, data&FlagPosZero)
		}
	}
	bus.Reset()
}

func TestRegisters(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x4)
	tempVal := uint64(0x6)
	flagsVal := FlagPosZero // Accumulator is empty to start

	alu.ReadFlags()
	verifyRegister(flagsVal, &bus, t)

	writeAccumulator(accumVal, &alu, &bus, t)
	writeTemp(tempVal, &alu, &bus, t)

	// Now read them back
	alu.ReadAccumulator()
	verifyRegister(accumVal, &bus, t)
	alu.ReadTemp()
	verifyRegister(tempVal, &bus, t)
}

func TestAdd(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x9)
	tempVal := uint64(0x6)
	flagsVal := FlagPosZero // Accumulator is empty to start

	verifyFlags(flagsVal, &alu, &bus, t)

	writeAccumulator(accumVal, &alu, &bus, t)
	writeTemp(tempVal, &alu, &bus, t)

	alu.SetMode(AluIntModeAdd)
	alu.Evaluate()
	alu.ReadEval()
	expVal := (accumVal + tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should not cause a carry
	flagsVal = 0
	verifyFlags(flagsVal, &alu, &bus, t)

	// Add a second time
	alu.Evaluate()
	alu.ReadEval()
	expVal = (accumVal + tempVal + tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should cause a carry
	flagsVal = FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestSub(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x9)
	tempVal := uint64(0x6)
	flagsVal := FlagPosZero // Accumulator is empty to start

	verifyFlags(flagsVal, &alu, &bus, t)

	writeAccumulator(accumVal, &alu, &bus, t)
	writeTemp(tempVal, &alu, &bus, t)

	alu.SetMode(AluIntModeSub)
	alu.Evaluate()
	alu.ReadEval()
	expVal := (accumVal - tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should not cause a borrow, which sets the carry bit
	flagsVal = FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)

	// Subtract a second time
	alu.Evaluate()
	alu.ReadEval()
	expVal = (accumVal - tempVal - tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should cause a borrow, which clears te carry bit
	flagsVal = 0
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestCLB(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x9)
	tempVal := uint64(0x9)
	flagsVal := FlagPosZero // Accumulator is empty to start

	writeAccumulator(accumVal, &alu, &bus, t)
	writeTemp(tempVal, &alu, &bus, t)

	alu.SetMode(AluIntModeAdd)
	alu.Evaluate()
	alu.ReadEval()
	expVal := (accumVal + tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should cause a carry
	flagsVal = FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)

	// Now clear both
	alu.ExectuteAccInst(CLB)

	flagsVal = FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)

	// Now read them back
	accumVal = 0
	alu.ReadAccumulator()
	verifyRegister(accumVal, &bus, t)
}

func TestCLC(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x9)
	tempVal := uint64(0x9)
	flagsVal := FlagPosZero // Accumulator is empty to start

	writeAccumulator(accumVal, &alu, &bus, t)
	writeTemp(tempVal, &alu, &bus, t)

	alu.SetMode(AluIntModeAdd)
	alu.Evaluate()
	alu.ReadEval()
	expVal := (accumVal + tempVal) & 0xF
	// Write this back to the accumulator
	alu.WriteAccumulator()

	verifyRegister(expVal, &bus, t)

	// This should cause a carry
	flagsVal = FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)

	// Now clear only the carry
	alu.ExectuteAccInst(CLC)

	flagsVal = 0
	alu.ReadFlags()
	verifyRegister(flagsVal, &bus, t)

	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)
}

func TestIAC(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0xE)

	writeAccumulator(accumVal, &alu, &bus, t)

	// Increment
	alu.ExectuteAccInst(IAC)

	expVal := accumVal + 1
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal := uint64(0) // No flags should be set
	verifyFlags(flagsVal, &alu, &bus, t)

	// The next incrment should wrap around to zero
	alu.ExectuteAccInst(IAC)

	expVal = 0
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	// Now we should have a carry and nothing in the accumulator
	flagsVal = FlagPosCarry | FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestCMC(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// We should have zero, and no carry to start with
	flagsVal := FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)

	alu.ExectuteAccInst(CMC)

	// Now the carry bit should be set
	flagsVal = FlagPosZero | FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)

	alu.ExectuteAccInst(CMC)

	// Now the carry bit should be unset
	flagsVal = FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)

}

func TestCMA(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0xA)

	writeAccumulator(accumVal, &alu, &bus, t)

	// Increment
	alu.ExectuteAccInst(CMA)

	expVal := uint64(0x5)
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)
}

func TestRAL(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	accumVal := uint64(0xA)
	writeAccumulator(accumVal, &alu, &bus, t)

	// Rotate
	alu.ExectuteAccInst(RAL)

	expVal := (accumVal << 1) & 0xf
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal := FlagPosCarry // The high bit should have rotated into the carry
	verifyFlags(flagsVal, &alu, &bus, t)

	// Rotate
	alu.ExectuteAccInst(RAL)

	expVal = (expVal<<1)&0xf + 1 // For the carry bit
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal = uint64(0) // The high bit should have rotated into the carry
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestRAR(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	accumVal := uint64(0x5)
	writeAccumulator(accumVal, &alu, &bus, t)

	// Rotate
	alu.ExectuteAccInst(RAR)

	expVal := (accumVal >> 1) & 0xf
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal := FlagPosCarry // The low bit should have rotated into the carry
	verifyFlags(flagsVal, &alu, &bus, t)

	// Rotate
	alu.ExectuteAccInst(RAR)

	expVal = (expVal>>1)&0xf + 8 // For the carry bit
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal = uint64(0) // The high bit should have rotated into the carry
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestTCC(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	accumVal := uint64(0x5)
	writeAccumulator(accumVal, &alu, &bus, t)

	// Carry is clear, so this should clear the accumulator
	alu.ExectuteAccInst(TCC)

	expVal := uint64(0)
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal := FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)

	// Set the carry bit, and try again
	alu.ExectuteAccInst(STC)

	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	flagsVal = FlagPosZero | FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestDAC(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	// The Alu registers are written and read through the bus
	accumVal := uint64(0x1)

	writeAccumulator(accumVal, &alu, &bus, t)

	// Decrement
	alu.ExectuteAccInst(DAC)

	expVal := accumVal - 1
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	// Carry here indicates no borrow
	flagsVal := FlagPosCarry | FlagPosZero
	verifyFlags(flagsVal, &alu, &bus, t)

	// Clear the carry bit
	alu.ExectuteAccInst(CLC)

	// The next decrment should wrap around to 0xF
	alu.ExectuteAccInst(DAC)

	expVal = 0xF
	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)

	// Now we should NOT have a carry (that indicates a borrow)
	flagsVal = 0
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestDAA(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	accumVal := uint64(0x1)
	writeAccumulator(accumVal, &alu, &bus, t)

	// This should have no effect
	alu.ExectuteAccInst(DAA)
	alu.ReadAccumulator()
	verifyRegister(accumVal, &bus, t)

	flagsVal := uint64(0)
	verifyFlags(flagsVal, &alu, &bus, t)

	accumVal = uint64(0x9)
	writeAccumulator(accumVal, &alu, &bus, t)
	alu.ExectuteAccInst(DAA)

	// Everything should stay the same
	alu.ReadAccumulator()
	verifyRegister(accumVal, &bus, t)
	verifyFlags(flagsVal, &alu, &bus, t)

	// If the accumulator has a value > 9, it will add 6
	accumVal = uint64(0xB)
	expVal := (accumVal + 6) & 0xf
	writeAccumulator(accumVal, &alu, &bus, t)
	alu.ExectuteAccInst(DAA)

	alu.ReadAccumulator()
	verifyRegister(expVal, &bus, t)
	// Now carry should be set
	flagsVal = FlagPosCarry
	verifyFlags(flagsVal, &alu, &bus, t)
}

func TestKBP(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	for i := 0; i < 15; i++ {
		writeAccumulator(uint64(i), &alu, &bus, t)

		switch i {
		case 0:
			fallthrough
		case 1:
			fallthrough
		case 2:
			// This should have no effect
			alu.ExectuteAccInst(KBP)
			alu.ReadAccumulator()
			verifyRegister(uint64(i), &bus, t)
		case 4:
			// This should convert 4 to 3
			alu.ExectuteAccInst(KBP)
			alu.ReadAccumulator()
			verifyRegister(uint64(3), &bus, t)
		case 8:
			// This should convert 8 to 4
			alu.ExectuteAccInst(KBP)
			alu.ReadAccumulator()
			verifyRegister(uint64(4), &bus, t)
		default:
			// This should set to 0xF for all other values (more than one bit set)
			alu.ExectuteAccInst(KBP)
			alu.ReadAccumulator()
			verifyRegister(uint64(0xF), &bus, t)
		}
	}
}

func TestDCL(t *testing.T) {
	SetupLogger()
	alu := Alu{}
	bus := common.Bus{}
	bus.Init(4, "Test Bus")
	alu.Init(&bus, 4)

	for i := 0; i < 15; i++ {
		writeAccumulator(uint64(i), &alu, &bus, t)
		alu.ExectuteAccInst(DCL)
		expRam := uint64(i & 7)
		gotRam := alu.GetCurrentRamBank()
		if expRam != gotRam {
			t.Errorf("RAM select mismatch. Exp %X, got %X", expRam, gotRam)
		}
	}
}
