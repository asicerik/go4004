package cpucore

import (
	"instruction"
	"os"
	"testing"

	"github.com/romana/rlog"
)

func SetupLogger() {
	// Programmatically change an rlog setting from within the program
	os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
	os.Setenv("RLOG_TRACE_LEVEL", "0")
	os.Setenv("RLOG_LOG_FILE", "cpucore_test.log")
	rlog.UpdateEnv()
}

func DumpState(core Core) {
	rlog.Infof("PC=%X, DBUS=%X, INST=%X, SYNC=%d, CCLK=%d",
		core.GetProgramCounter(),
		core.ExternalDataBus.Read(),
		core.GetInstructionRegister(),
		core.Sync, core.GetClockCount())
}

func TestSync(t *testing.T) {
	// SetupLogger()
	core := Core{}
	core.Init()
	syncSeen, count := waitForSync(&core)
	if !syncSeen {
		t.Error("Sync was not seen")
	}
	if count != 7 {
		t.Errorf("Count was not 7, it was %d", count)
	}
}

func waitForSync(core *Core) (syncSeen bool, count int) {
	for i := 0; i < 16; i++ {
		core.Calculate()
		core.ClockIn()
		core.ClockOut()
		if core.Sync == 0 {
			if !syncSeen {
				syncSeen = true
				count = 0
			} else {
				// Run 1 extra clock to align to the start
				core.Calculate()
				core.ClockIn()
				core.ClockOut()
				break
			}
		} else if syncSeen {
			count++
		}
	}
	return
}

func runOneCycle(core *Core, data uint64, t *testing.T) (addr uint64) {
	addr, _ = runOneIOCycle(core, data, t)
	return addr
}

func runOneIOCycle(core *Core, data uint64, t *testing.T) (addr uint64, ioVal uint64) {
	addr = 0
	for i := 0; i < 8; i++ {
		DumpState(*core)
		core.Calculate()
		core.ClockIn()
		if i < 3 {
			addr = addr | (core.ExternalDataBus.Read() << (uint64(i) * 4))
		}
		if i == 6 {
			ioVal = core.ExternalDataBus.Read()
		}
		core.ClockOut()
		if i == 2 {
			rlog.Debugf("runOneCycle: Writing upper data %X", (data>>4)&0xf)
			core.ExternalDataBus.Write((data >> 4) & 0xf)
		} else if i == 3 {
			rlog.Debugf("runOneCycle: Writing lower data %X", data&0xf)
			core.ExternalDataBus.Write(data & 0xf)
		}
	}
	return
}

func TestProgramCounterBasic(t *testing.T) {
	// SetupLogger()
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}

	nextAddr := uint64(1) // since we already ran the first cycle
	// run 4 complete cycles
	for i := 0; i < 4; i++ {
		addr := runOneCycle(&core, 0, t)
		if addr != nextAddr {
			t.Errorf("Address %X was not equal to %X", addr, i)
		}
		nextAddr++
	}
}

func TestJUN(t *testing.T) {
	SetupLogger()
	rlog.Info("TestJUN")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	verifyJump(&core, instruction.JUN, true, t)
}

func TestJCN(t *testing.T) {
	SetupLogger()
	rlog.Info("TestJCN")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	// No flags set, should jump
	conditionFlags := uint64(0)
	jumpExpected := true
	verifyJump(&core, instruction.JCN|conditionFlags, jumpExpected, t)

	// Carry bit should not be set, no jump
	conditionFlags = uint64(2)
	jumpExpected = false
	verifyJump(&core, instruction.JCN|conditionFlags, jumpExpected, t)

	// Accumulator bit should be set, jump
	conditionFlags = uint64(4)
	jumpExpected = true
	verifyJump(&core, instruction.JCN|conditionFlags, jumpExpected, t)

	// Load the accumulator and verify no jump
	runOneCycle(&core, instruction.LDM|5, t)
	conditionFlags = uint64(4)
	jumpExpected = false
	verifyJump(&core, instruction.JCN|conditionFlags, jumpExpected, t)

	// Run the inverse test
	runOneCycle(&core, instruction.LDM|5, t)
	conditionFlags = uint64(0xC)
	jumpExpected = true
	verifyJump(&core, instruction.JCN|conditionFlags, jumpExpected, t)

	// TODO: Carry Tests
}

func verifyJump(core *Core, instruction uint64, jumpExpected bool, t *testing.T) {
	addr := uint64(0)
	data := uint64(0)
	nextAddr := runOneCycle(core, 0, t) + 1
	// run 5 complete cycles
	for i := 0; i < 5; i++ {
		switch i {
		case 0:
		case 4:
			fallthrough
		case 1:
			data = 0x0
		case 2:
			data = instruction
		case 3:
			data = 0xaa
		}
		addr = runOneCycle(core, data, t)
		if addr != nextAddr {
			t.Errorf("Address %X was not equal to %X", addr, nextAddr)
		}
		if i == 3 && jumpExpected {
			nextAddr = data
		} else {
			nextAddr++
		}
	}

}

func loadRegisterPair(core *Core, data uint8, regPair int, t *testing.T) (nextAddr uint64) {
	// Load the accumulator with the lower 4 bits
	nextAddr = runOneCycle(core, uint64(instruction.LDM|(data&0xf)), t)
	// Swap the accumulator with the lower register pair
	nextAddr = runOneCycle(core, uint64(instruction.XCH|(regPair<<1)), t)
	// Load the accumulator with the higher 4 bits
	nextAddr = runOneCycle(core, uint64(instruction.LDM|((data>>4)&0xf)), t)
	// Swap the accumulator with the lower register pair
	nextAddr = runOneCycle(core, uint64(instruction.XCH|((regPair<<1)+1)), t)
	return nextAddr
}

func TestSRC(t *testing.T) {
	SetupLogger()
	rlog.Info("TestSRC")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	regPair := 2
	expSrcVal := uint64(0xd)
	// Populate the scratch registers with out expected value
	loadRegisterPair(&core, 0xd, regPair, t)
	core.LogScratchPadRegisters()

	// Run the SRC command
	_, srcVal := runOneIOCycle(&core, uint64(instruction.SRC|(regPair<<1)), t)
	if expSrcVal != srcVal {
		t.Errorf("SRC val %X was not equal to %X", srcVal, expSrcVal)
	}

}

func TestFIM(t *testing.T) {
	SetupLogger()
	rlog.Info("TestFIM")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	regPair := 2
	romValue := 0xde

	// The first cycle sets up the register pair to load into
	runOneCycle(&core, uint64(instruction.FIM|(regPair<<1)), t)
	// The second cycle provides the data to load
	runOneCycle(&core, uint64(romValue), t)

	core.LogScratchPadRegisters()
}

func TestFIN(t *testing.T) {
	SetupLogger()
	rlog.Info("TestFIN")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	regPair := 2
	romAddr := uint8(0xde)
	romData := uint8(0x77)

	// Populate scratch registers pair 0 with out expected address
	loadRegisterPair(&core, romAddr, 0, t)
	core.LogScratchPadRegisters()

	// Run the command and verify the address on the next cycle
	addr := runOneCycle(&core, uint64(instruction.FIN|(regPair<<1)), t)

	// When the fetch is done, we should resume where we left off
	expAddr := addr + 1

	// Run the next cycle and provide the ROM read data
	addr = runOneCycle(&core, uint64(romData), t)
	if addr != uint64(romAddr) {
		t.Errorf("Fetch address mismatch. Exp %X, got %X", romAddr, addr)
	}

	// Run a final cycle to see where the program counter ended up
	addr = runOneCycle(&core, uint64(romData), t)
	if addr != uint64(expAddr) {
		t.Errorf("Continue address mismatch. Exp %X, got %X", expAddr, addr)
	}
}

func TestJIN(t *testing.T) {
	SetupLogger()
	rlog.Info("TestJIN")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	regPair := 2
	romAddr := uint8(0xde)

	// Populate the scratch registers with out expected value
	loadRegisterPair(&core, romAddr, regPair, t)
	core.LogScratchPadRegisters()

	// Run the command and verify the address on the next cycle
	runOneCycle(&core, uint64(instruction.JIN|(regPair<<1)), t)

	addr := runOneCycle(&core, uint64(instruction.NOP), t)
	if addr != uint64(romAddr) {
		t.Errorf("Address %X was not equal to %X", addr, romAddr)
	}
}
