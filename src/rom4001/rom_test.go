package rom4001

import (
	"common"
	"os"
	"testing"

	"github.com/romana/rlog"
)

func SetupLogger() {
	// Programmatically change an rlog setting from within the program
	os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
	os.Setenv("RLOG_TRACE_LEVEL", "0")
	os.Setenv("RLOG_LOG_FILE", "rom_test.log")
	rlog.UpdateEnv()
}

type romTestJig struct {
	rom     Rom4001
	dataBus common.Bus
	sync    int
}

func DumpState(jig *romTestJig) {
	rlog.Infof("DBUS=%X, SYNC=%d, CCLK=%d",
		jig.dataBus.Read(),
		jig.sync,
		jig.rom.GetClockCount())
}

func syncROM(jig *romTestJig) {
	for i := 0; i < 8; i++ {
		if (i % 8) == 7 {
			jig.sync = 0
		} else {
			jig.sync = 1
		}
		//DumpState(jig)
		jig.rom.ClockIn()
		jig.rom.ClockOut()
		jig.dataBus.Reset()
		jig.dataBus.Write(0)
	}
}

func readROM(jig *romTestJig, addr uint64) (data uint64) {
	for i := 0; i < 8; i++ {
		// Write to ROM block
		jig.sync = 1
		jig.dataBus.Reset()
		switch i {
		case 0:
			jig.dataBus.Write(addr & 0xf)
		case 1:
			jig.dataBus.Write((addr >> 4) & 0xf)
		case 2:
			jig.dataBus.Write((addr >> 8) & 0xf)
		case 7:
			jig.sync = 0
		}
		DumpState(jig)
		jig.rom.ClockIn()
		jig.dataBus.Reset()
		jig.rom.ClockOut()
		// Read from ROM block
		switch i {
		case 2:
			data = jig.dataBus.Read() << 4
		case 3:
			data = data | jig.dataBus.Read()
			rlog.Debugf("Read %02X from ROM", data)
		}
	}
	return
}

func TestSync(t *testing.T) {
	SetupLogger()
	jig := romTestJig{}
	jig.dataBus.Init(4, "JIG external bus")
	jig.rom.Init(&jig.dataBus, &jig.sync)
	syncROM(&jig)
	readROM(&jig, 0x012)
	readROM(&jig, 0x000)
}
