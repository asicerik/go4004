package rom4001

import (
	"common"
	"instruction"
	"os"
	"testing"

	"github.com/romana/rlog"
)

func SetupLogger() {
	// Programmatically change an rlog setting from within the program
	os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
	//os.Setenv("RLOG_TRACE_LEVEL", "0")
	os.Setenv("RLOG_LOG_FILE", "rom_test.log")
	rlog.UpdateEnv()
	rlog.Info("Test starting ***********************")
}

type romTestJig struct {
	rom     Rom4001
	dataBus common.Bus
	ioBus   common.Bus
	sync    int
	cmRom   int
}

func createTestJig() *romTestJig {
	jig := romTestJig{}
	jig.dataBus.Init(4, "JIG external bus")
	jig.ioBus.Init(4, "ROM I/O Bus")
	jig.rom.Init(&jig.dataBus, &jig.sync, &jig.cmRom)
	jig.rom.SetIOBus(&jig.ioBus)
	return &jig
}

func DumpState(jig *romTestJig) {
	rlog.Infof("DBUS=%X, IOBUS=%X, SYNC=%d, CCLK=%d",
		jig.dataBus.Read(),
		jig.ioBus.Read(),
		jig.sync,
		jig.rom.GetClockCount())
}

func generateBlankROMImage() []uint8 {
	data := make([]uint8, Depth)
	return data
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
func readROM(jig *romTestJig, addr uint64) uint8 {
	return readROMFull(jig, addr, nil, false)
}

func readROMFull(jig *romTestJig, addr uint64, ioData *uint64, ioRead bool) uint8 {
	var data uint64
	for i := 0; i < 8; i++ {
		// Write to ROM block
		jig.sync = 1
		switch i {
		case 0:
			jig.dataBus.Reset()
			jig.dataBus.Write(addr & 0xf)
		case 1:
			jig.dataBus.Reset()
			jig.dataBus.Write((addr >> 4) & 0xf)
		case 2:
			jig.dataBus.Reset()
			jig.dataBus.Write((addr >> 8) & 0xf)
		case 6:
			if !ioRead && ioData != nil {
				jig.dataBus.Reset()
				jig.dataBus.Write(*ioData & 0xf)
			}
		case 7:
			jig.sync = 0
		}
		DumpState(jig)
		jig.rom.ClockIn()
		jig.dataBus.Reset()
		jig.rom.ClockOut()
		// Read from ROM block
		// NOTE: these indicies are one earlier than the actual clock cycle number
		switch i {
		case 2:
			data = jig.dataBus.Read() << 4
		case 3:
			data = data | jig.dataBus.Read()
			rlog.Debugf("Read %02X from ROM", data)
		case 5:
			if ioRead {
				data = jig.dataBus.Read()
			}
		}
	}
	return uint8(data)
}

func TestDataRead(t *testing.T) {
	SetupLogger()
	jig := createTestJig()
	romImage := generateBlankROMImage()
	jig.rom.LoadProgram(romImage)

	syncROM(jig)

	// Verify we can read the correct data
	romImage[0x12] = 0xde
	romImage[0xaa] = 0xad
	jig.rom.LoadProgram(romImage)
	data := readROM(jig, 0x0)
	if data != romImage[0x0] {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", romImage[0x0], data)
	}
	data = readROM(jig, 0x012)
	if data != romImage[0x12] {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", romImage[0x12], data)
	}
	data = readROM(jig, 0x0aa)
	if data != romImage[0xaa] {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", romImage[0xAA], data)
	}
}

func TestChipSelectRead(t *testing.T) {
	SetupLogger()
	jig := createTestJig()
	romImage := generateBlankROMImage()
	jig.rom.LoadProgram(romImage)

	syncROM(jig)

	// Verify we can read the correct data
	romImage[0x12] = 0xde
	romImage[0xaa] = 0xad
	jig.rom.LoadProgram(romImage)

	// Read with CM_ROM enabled
	jig.cmRom = 0 // active low
	data := readROM(jig, 0x012)
	if data != romImage[0x12] {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", romImage[0x12], data)
	}

	// Read with CM_ROM disabled
	jig.cmRom = 1 // active low
	data = readROM(jig, 0x012)
	if data != 0xFF {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", 0xFF, data)
	}

	// Read with CM_ROM enabled, but with an address outside of our ROM
	jig.cmRom = 0 // active low
	data = readROM(jig, 0x512)
	if data != 0xFF {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", 0xFF, data)
	}

	// Finally, verify we can change our chip ID to match this address
	jig.cmRom = 0 // active low
	jig.rom.SetChipID(5)
	data = readROM(jig, 0x512)
	if data != romImage[0x12] {
		t.Errorf("ROM read data mismatch. exp %02X, got %02X", romImage[0x12], data)
	}
}

func TestIOWrite(t *testing.T) {
	SetupLogger()
	jig := createTestJig()
	romImage := generateBlankROMImage()
	// Setup an I/O write program
	romImage[0] = instruction.FIM_SRC | 1 // Mark this is a SRC
	romImage[1] = instruction.WRR         // ROM I/O write
	romImage[2] = instruction.FIM_SRC | 1 // Mark this is a SRC
	romImage[3] = instruction.RDR         // ROM I/O read
	jig.rom.LoadProgram(romImage)

	syncROM(jig)

	// ROM I/O write
	ioData := uint64(0) // Select chip 0
	readROMFull(jig, 0, &ioData, false)
	ioData = 0xC // write 0xC to the IO port on the ROM
	readROMFull(jig, 1, &ioData, false)
	// We should now see the data on the IO bus
	if ioData != jig.ioBus.Read() {
		t.Errorf("I/O bus did not match. Exp %X, got %X", ioData, jig.ioBus.Read())
	}

	// ROM I/O read
	ioData = uint64(0) // Select chip 0
	readROMFull(jig, 2, &ioData, false)
	// Set the I/O bus to our expected value
	jig.ioBus.Reset()
	jig.ioBus.Write(0xA)
	ioData = uint64(readROMFull(jig, 3, &ioData, true))
	// We should now see the data on the IO bus
	if ioData != jig.ioBus.Read() {
		t.Errorf("I/O bus did not match. Exp %X, got %X", ioData, jig.ioBus.Read())
	}

}
