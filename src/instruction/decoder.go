package instruction

import (
	"common"

	"github.com/romana/rlog"
)

const JUN = 0x40 // Jump unconditional
const LDM = 0xD0 // Load direct into accumulator
const XCH = 0xB0 // Exchange the accumulator and scratchpad register

type DecoderFlag struct {
	Name    string
	Value   int
	Changed bool
}

type Decoder struct {
	// Control Flags
	Flags map[int]DecoderFlag

	clockCount      int
	syncSent        bool
	currInstruction int
}

const (
	// External I/O Comes first in the list
	Sync              = iota // We should output the SYNC signal
	BusDir                   // External data bus direction
	BusTurnAround            // If true, swap the bus direction after the first action
	InstRegOut               // Instruction register should drive the bus
	InstRegLoad              // Load the instruction register i/o buffer
	PCOut                    // Program Counter should drive the bus
	PCLoad                   // Load the program counter from the internal bus
	PCInc                    // Increment the program counter
	AccOut                   // Accumulator register should drive the bus
	AccLoad                  // Load the accumulator from the internal bus
	TempLoad                 // Load the temp register from the internal bus
	TempOut                  // Temp register should drive the bus
	ScratchPadIndex          // Which scratchpd (index) register to read/write
	ScratchPadLoad4          // Load 4 bits into the currently selected scratchpad register
	ScratchPadLoad8          // Load 8 bits into the currently selected scratchpad registers
	ScratchPadOut            // Currently selected scratchpad register should drive the bus
	DecodeInstruction        // The instruction register is ready to be decoded
	END                      // Marker for end of list
)

func (d *Decoder) Init() {
	d.clockCount = 0
	d.syncSent = false
	d.currInstruction = -1
	d.Flags = make(map[int]DecoderFlag)
	d.Flags[Sync] = DecoderFlag{"SYNC", 0, false}
	d.Flags[BusDir] = DecoderFlag{"BDIR", 0, false}
	d.Flags[BusTurnAround] = DecoderFlag{"BTA ", 0, false}
	d.Flags[InstRegOut] = DecoderFlag{"IO  ", 0, false}
	d.Flags[InstRegLoad] = DecoderFlag{"IL  ", 0, false}
	d.Flags[PCOut] = DecoderFlag{"PCO ", 0, false}
	d.Flags[PCLoad] = DecoderFlag{"PCL ", 0, false}
	d.Flags[PCInc] = DecoderFlag{"PCI ", 0, false}
	d.Flags[AccOut] = DecoderFlag{"AO  ", 0, false}
	d.Flags[AccLoad] = DecoderFlag{"AL  ", 0, false}
	d.Flags[TempOut] = DecoderFlag{"TO  ", 0, false}
	d.Flags[TempLoad] = DecoderFlag{"TL  ", 0, false}
	d.Flags[ScratchPadIndex] = DecoderFlag{"SPI ", 0, false}
	d.Flags[ScratchPadLoad4] = DecoderFlag{"SPL4", 0, false}
	d.Flags[ScratchPadLoad8] = DecoderFlag{"SPL8", 0, false}
	d.Flags[ScratchPadOut] = DecoderFlag{"SPO ", 0, false}
	d.Flags[DecodeInstruction] = DecoderFlag{"DEC ", 0, false}
}

func (d *Decoder) GetClockCount() int {
	return d.clockCount
}

func (d *Decoder) resetFlags() {
	for i := 0; i < END; i++ {
		d.clearFlag(i, 0)
	}
}

func (d *Decoder) clearFlag(index int, value int) {
	flag := d.Flags[index]
	flag.Changed = flag.Value != value
	flag.Value = value
	d.Flags[index] = flag
}

func (d *Decoder) writeFlag(index int, value int) {
	flag := d.Flags[index]
	flag.Changed = true // we always set changed so the UI can show the write
	flag.Value = value
	d.Flags[index] = flag
	rlog.Tracef(0, "Wrote Flag: Name=%s, value=%d. ClkCnt=%d", flag.Name, d.Flags[index].Value, d.clockCount)
}

func (d *Decoder) Clock() {
	d.resetFlags()

	if d.clockCount < 7 {
		d.clockCount++
	} else {
		d.clockCount = 0
		// Handle startup condition. Make sure we send sync before incrementing
		// the program counter
		if d.syncSent {
			d.writeFlag(PCInc, 1)
		}
		d.syncSent = true // FIXME
	}

	// Defaults

	switch d.clockCount {
	case 0:
		// Drive the current address (nybble 0) to the external bus
		d.writeFlag(BusDir, common.DirOut)
		d.writeFlag(PCOut, 1)
	case 1:
		// Drive the current address (nybble 1) to the external bus
		d.writeFlag(BusDir, common.DirOut)
		d.writeFlag(PCOut, 1)
	case 2:
		// Drive the current address (nybble 2) to the external bus
		d.writeFlag(BusDir, common.DirOut)
		d.writeFlag(PCOut, 1)
	case 4:
		fallthrough
	case 5:
		if d.syncSent {
			// Read the OPR from the external bus and write it into the instruction register
			d.writeFlag(BusDir, common.DirIn)
			d.writeFlag(InstRegLoad, 1)
		}
		if d.clockCount == 5 {
			d.writeFlag(DecodeInstruction, 1)
		}
	case 7:
		d.writeFlag(Sync, 1)
	}
	// Continue to decode instructions after clock 5
	if d.clockCount != 5 && d.currInstruction > 0 {
		d.decodeCurrentInstruction()
	}
}

// SetCurrentInstruction set the current instruction from the instruction register
func (d *Decoder) SetCurrentInstruction(inst uint64) (err error) {
	if inst != 0 {
		rlog.Debugf("SetCurrentInstruction: %02X", inst)
	}
	d.currInstruction = int(inst)
	return d.decodeCurrentInstruction()
}

func (d *Decoder) decodeCurrentInstruction() (err error) {

	switch d.currInstruction & 0xf0 {
	// 	// This is special case code in the core as we will need to turn the bus around
	// 	if d.clockCount == 5 {
	// 		d.InstRegOut = true
	// 		d.BusTurnAround = true
	// 	}
	// }
	case JUN:
		rlog.Debug("JUN command decoded")
		// FIXME : Need to load the address
		d.writeFlag(ScratchPadIndex, 0)
		d.writeFlag(ScratchPadOut, 1)
		d.writeFlag(PCLoad, 1)
	case XCH:
		rlog.Debug("XCH command decoded")
		// Exchange the accumulator and the scratchpad register
		if d.clockCount == 5 {
			// Load the current scratchpad register into the temp register
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
			d.writeFlag(ScratchPadOut, 1)
			d.writeFlag(TempLoad, 1)
		} else if d.clockCount == 6 {
			// Load the accumulator into the current scratchpad register
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
			d.writeFlag(ScratchPadLoad4, 1)
			d.writeFlag(AccOut, 1)
		} else if d.clockCount == 7 {
			// Load the temp register into the accumulator
			d.writeFlag(TempOut, 1)
			d.writeFlag(AccLoad, 1)
			d.currInstruction = -1
		}
	case LDM:
		rlog.Debug("LDM command decoded")
		d.writeFlag(AccLoad, 1)
		d.writeFlag(InstRegOut, 1)
		d.currInstruction = -1
	}
	return err
}
