package instruction

import (
	"common"

	"github.com/romana/rlog"
)

const NOP = 0x00     // No Operation
const JCN = 0x10     // Jump conditional
const FIM_SRC = 0x20 // FIM = Fetch immediate. SRC = Send address to ROM/RAM
const JUN = 0x40     // Jump unconditional
const LDM = 0xD0     // Load direct into accumulator
const XCH = 0xB0     // Exchange the accumulator and scratchpad register
const WRR = 0xE2     // ROM I/O write
const RDR = 0xEA     // ROM I/O read

type DecoderFlag struct {
	Name    string
	Value   int
	Changed bool
}

type Decoder struct {
	// Control Flags
	Flags map[int]DecoderFlag

	clockCount      int // Internal clock count
	instPhase       int // which instruction phase are we in
	syncSent        bool
	currInstruction int
	dblInstruction  int  // The current instruction requires two complete cycles
	inhibitPCInc    bool // Inhibit the program counter increment for jumps, etc.
	x2IsRead        bool // The CPU's X2 cycle is an external device read
	x3IsRead        bool // The CPU's X3 cycle is an external device read
}

const (
	// External I/O Comes first in the list
	Sync              = iota // We should output the SYNC signal
	BusDir                   // External data bus direction
	BusTurnAround            // If true, swap the bus direction after the first action
	InstRegOut               // Instruction register should drive the bus (value is the nybble to load)
	InstRegLoad              // Load the instruction register i/o buffer
	PCOut                    // Program Counter should drive the bus
	PCLoad                   // Load the program counter from the internal bus (value is the nybble to load)
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
	EvalulateJCN             // Evaluate the condition flags for a JCN instruction
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
	d.Flags[EvalulateJCN] = DecoderFlag{"EJCN", 0, false}
}

func (d *Decoder) GetClockCount() int {
	return d.instPhase
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

// Clock updates flip flops on rising edge of the clock
func (d *Decoder) Clock() {
	d.instPhase = d.clockCount // delayed by one clock
	if d.clockCount < 7 {
		d.clockCount++
	} else {
		d.clockCount = 0
	}
	if d.Flags[Sync].Value == 1 {
		d.syncSent = true
	}
}

func (d *Decoder) CalculateFlags() {
	d.resetFlags()

	if d.clockCount == 7 {
		// Handle startup condition. Make sure we send sync before incrementing
		// the program counter
		if d.syncSent {
			if !d.inhibitPCInc {
				d.writeFlag(PCInc, 1)
			}
		}
	}

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
		if d.syncSent {
			// Read the OPR from the external bus and write it into the instruction register
			d.writeFlag(BusDir, common.DirIn)
			if d.clockCount == 4 {
				d.writeFlag(InstRegLoad, 1)
			}
		}
	case 5:
		if d.syncSent {
			d.writeFlag(BusDir, common.DirIn)
			d.writeFlag(InstRegLoad, 1)
			d.writeFlag(DecodeInstruction, 1)
		}
	case 6:
		if d.syncSent {
			// If the X2 cycle is a read, read the external bus
			if d.x2IsRead {
				d.writeFlag(BusDir, common.DirIn)
			} else {
				d.writeFlag(BusDir, common.DirOut)
			}
		}
	case 7:
		// If the X3 cycle is a read, read the external bus
		if d.x3IsRead {
			d.writeFlag(BusDir, common.DirIn)
		}
		d.writeFlag(Sync, 1)
	}
	// Continue to decode instructions after clock 5
	if d.clockCount != 5 && d.currInstruction > 0 {
		d.decodeCurrentInstruction(false)
	}
}

// SetCurrentInstruction set the current instruction from the instruction register
func (d *Decoder) SetCurrentInstruction(inst uint64, evalResult bool) (err error) {
	if d.dblInstruction == 0 {
		if inst != 0 {
			rlog.Debugf("SetCurrentInstruction: %02X", inst)
		}
		d.currInstruction = int(inst)
	} else {
		d.currInstruction = d.dblInstruction
	}
	return d.decodeCurrentInstruction(evalResult)
}

func (d *Decoder) decodeCurrentInstruction(evalResult bool) (err error) {
	// The upper 4 bits of the instruction
	opr := d.currInstruction & 0xf0
	fullInst := d.currInstruction
	switch opr {
	case JCN:
		fallthrough
	case JUN:
		// Are we on the first phase?
		if d.dblInstruction == 0 {
			if opr == JCN {
				rlog.Debug("JCN command decoded")
				d.writeFlag(EvalulateJCN, 1)
			} else if opr == JUN {
				rlog.Debug("JUN command decoded")
			}
			if d.clockCount == 5 {
				d.writeFlag(InstRegOut, 1)
			} else if d.clockCount == 6 {
				// Store the upper 4 bits in the temp register
				d.writeFlag(TempLoad, 1)
				d.dblInstruction = d.currInstruction
				d.currInstruction = -1
			}
		} else {
			if d.clockCount == 5 {
				// If this is a conditional jump, evaluate the condition here
				blockJump := false
				if opr == JCN {
					blockJump = !evalResult
				}
				if !blockJump {
					// Block the PC increment
					d.inhibitPCInc = true
					// Output the lower 4 bits of the instruction register
					// It contains the lowest 4 bits of the address
					d.writeFlag(InstRegOut, 1)
				} else {
					rlog.Info("Conditional jump was not taken")
					d.dblInstruction = 0
					d.currInstruction = -1
				}
			} else if d.clockCount == 6 {
				// Load the lowest 4 bits into the PC
				d.writeFlag(PCLoad, 1)
				// Output the higher 4 bits of the instruction register
				// It contains the middle 4 bits of the address
				d.writeFlag(InstRegOut, 2)
			} else if d.clockCount == 7 {
				// Load the middle 4 bits into the PC
				d.writeFlag(PCLoad, 2)
				// Output the temp register
				// It contains the middle 4 bits of the address
				d.writeFlag(TempOut, 1)
			} else if d.clockCount == 0 {
				if opr == JUN {
					// NOTE: we have already started outputting the PC onto the bus
					// for the next cycle, but we can still update the highest bits
					// since they go out last
					// Load the highest 4 bits into the PC
					d.writeFlag(PCLoad, 3)
				}
				d.dblInstruction = 0
				d.currInstruction = -1
				// Unblock the PC increment
				d.inhibitPCInc = false
			}
		}

	case XCH:
		rlog.Debug("XCH command decoded")
		// Exchange the accumulator and the scratchpad register
		if d.clockCount == 5 {
			// Output the current scratchpad register
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
			d.writeFlag(ScratchPadOut, 1)
		} else if d.clockCount == 6 {
			// Load the data from the previous into the Temp register
			d.writeFlag(TempLoad, 1)

			// Output the accumulator
			d.writeFlag(AccOut, 1)
		} else if d.clockCount == 7 {
			// Load the data from the previous into the scratchpad register
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
			d.writeFlag(ScratchPadLoad4, 1)

			// Output the temp register
			d.writeFlag(TempOut, 1)
		} else if d.clockCount == 0 {
			// Load the data from the previous into the accumulator register
			d.writeFlag(AccLoad, 1)
			d.currInstruction = -1
		}
	case LDM:
		rlog.Debug("LDM command decoded")
		if d.clockCount == 5 {
			d.writeFlag(InstRegOut, 1)
		} else if d.clockCount == 6 {
			d.writeFlag(AccLoad, 1)
			d.currInstruction = -1
		}
	case FIM_SRC:
		if (fullInst & 0x1) == 0 {
			rlog.Debug("FIM command decoded")
		} else {
			// Send I/O address to ROM/RAM
			rlog.Debug("SRC command decoded")
			if d.clockCount == 6 {
				// Output the selected scratchpad register
				d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)) // Note - we are chopping bit 0
				d.writeFlag(ScratchPadOut, 1)
			} else if d.clockCount == 7 {
				// Output the selected scratchpad register + 1
				d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)+1) // Note - we are chopping bit 0
				d.writeFlag(ScratchPadOut, 1)
				d.currInstruction = -1
			}
		}
	}
	switch fullInst {
	case WRR:
		// ROM I/O Write
		// Send I/O address to ROM/RAM
		rlog.Debug("WRR command decoded")
		if d.clockCount == 6 {
			// Output the accumulator to the external bus
			d.writeFlag(AccOut, 1)
			d.currInstruction = -1
		}
	}
	return err
}
