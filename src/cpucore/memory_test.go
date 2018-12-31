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
	addr = 0
	for i := 0; i < 8; i++ {
		DumpState(*core)
		core.Calculate()
		core.ClockIn()
		if i < 3 {
			addr = addr | (core.ExternalDataBus.Read() << (uint64(i) * 4))
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
	// SetupLogger()
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
