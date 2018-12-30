package instruction

import (
	"common"

	"github.com/romana/rlog"
)

const JUN = 0x40 // Jump unconditional
const LDM = 0xD0 // Load direct into accumulator
const XCH = 0xB0 // Exchange the accumulator and scratchpad register

type Decoder struct {
	// Control Flags
	BusDir            int  // External data bus direction
	BusTurnAround     bool // If true, swap the bus direction after the first action
	InstRegOut        bool // Instruction register should drive the bus
	InstRegLoad       bool // Load the instruction register i/o buffer
	PCOut             bool // Program Counter should drive the bus
	PCLoad            bool // Load the program counter from the internal bus
	PCInc             bool // Increment the program counter
	AccLoad           bool // Load the accumulator from the internal bus
	AccOut            bool // Accumulator register should drive the bus
	TempLoad          bool // Load the temp register from the internal bus
	TempOut           bool // Temp register should drive the bus
	ScratchPadLoad4   bool // Load 4 bits into the currently selected scratchpad register
	ScratchPadLoad8   bool // Load 8 bits into the currently selected scratchpad registers
	ScratchPadOut     bool // Currently selected scratchpad register should drive the bus
	DecodeInstruction bool // The instruction register is ready to be decoded
	Sync              bool // We should output the SYNC signal

	ScratchPadIndex int
	clockCount      int
	syncSent        bool
	currInstruction int
}

func (d *Decoder) Init() {
	d.clockCount = 0
	d.syncSent = false
	d.currInstruction = -1
}

func (d *Decoder) GetClockCount() int {
	return d.clockCount
}

func (d *Decoder) resetFlags() {
	d.BusDir = common.DirNone
	d.BusTurnAround = false
	d.InstRegOut = false
	d.InstRegLoad = false
	d.PCOut = false
	d.PCLoad = false
	d.PCInc = false
	d.AccLoad = false
	d.AccOut = false
	d.TempLoad = false
	d.TempOut = false
	d.ScratchPadLoad4 = false
	d.ScratchPadLoad8 = false
	d.ScratchPadOut = false
	d.DecodeInstruction = false
	d.Sync = false
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
			d.PCInc = true
		}
		d.syncSent = true // FIXME
	}

	// Defaults

	switch d.clockCount {
	case 0:
		// Drive the current address (nybble 0) to the external bus
		d.BusDir = common.DirOut
		d.PCOut = true
	case 1:
		// Drive the current address (nybble 1) to the external bus
		d.BusDir = common.DirOut
		d.PCOut = true
	case 2:
		// Drive the current address (nybble 2) to the external bus
		d.BusDir = common.DirOut
		d.PCOut = true
	case 4:
		fallthrough
	case 5:
		if d.syncSent {
			// Read the OPR from the external bus and write it into the instruction register
			d.BusDir = common.DirIn
			d.InstRegLoad = true
		}
		if d.clockCount == 5 {
			d.DecodeInstruction = true
		}
	case 7:
		d.Sync = true
	}
	// Continue to decode instructions after clock 5
	if d.clockCount != 5 && d.currInstruction > 0 {
		d.decodeCurrentInstruction()
	}
}

// SetCurrentInstruction set the current instruction from the instruction register
func (d *Decoder) SetCurrentInstruction(inst uint64) (err error) {
	rlog.Debugf("SetCurrentInstruction: %02X", inst)
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
		d.ScratchPadIndex = 0
		d.ScratchPadOut = true
		d.PCLoad = true
	case XCH:
		rlog.Debug("XCH command decoded")
		// Exchange the accumulator and the scratchpad register
		if d.clockCount == 5 {
			// Load the current scratchpad register into the temp register
			d.ScratchPadIndex = int(d.currInstruction & 0xf)
			d.ScratchPadOut = true
			d.TempLoad = true
		} else if d.clockCount == 6 {
			// Load the accumulator into the current scratchpad register
			d.ScratchPadIndex = int(d.currInstruction & 0xf)
			d.ScratchPadLoad4 = true
			d.AccOut = true
		} else if d.clockCount == 7 {
			// Load the temp register into the accumulator
			d.TempOut = true
			d.AccLoad = true
			d.currInstruction = -1
		}
	case LDM:
		rlog.Debug("LDM command decoded")
		d.AccLoad = true
		d.InstRegOut = true
		d.currInstruction = -1
	}
	return err
}
