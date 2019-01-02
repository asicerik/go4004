package instruction

import (
	"common"

	"github.com/romana/rlog"
)

const NOP = 0x00 // No Operation
const JCN = 0x10 // Jump conditional
const FIM = 0x20 // Fetch immediate
const SRC = 0x21 // Send address to ROM/RAM
const FIN = 0x30 // Fetch indirect from ROM
const JIN = 0x31 // Jump indirect from current register pair
const JUN = 0x40 // Jump unconditional
const JMS = 0x50 // Jump to subroutine
const INC = 0x60 // Increment register
const ISZ = 0x70 // Increment register and jump if zero
const ADD = 0x80 // Add register to accumulator with carry
const SUB = 0x90 // Subtract register from accumulator with borrow
const LDM = 0xD0 // Load direct into accumulator
const LD = 0xA0  // Load register into accumulator
const XCH = 0xB0 // Exchange the accumulator and scratchpad register
const BBL = 0xC0 // Branch back (stack pop)
const WRR = 0xE2 // ROM I/O write
const RDR = 0xEA // ROM I/O read
const ACC = 0xF0 // Alias for all the accumulator instructions

// Some helpers
// You can use any of these three in combination
const JCN_TEST_SET = 0x11  // Jump if test bit is set
const JCN_CARRY_SET = 0x12 // Jump if carry bit is set
const JCN_ZERO_SET = 0x14  // Jump if accumulator is zero
// You can use any of these three in combination
const JCN_TEST_UNSET = 0x19  // Jump if test bit is NOT set
const JCN_CARRY_UNSET = 0x1A // Jump if carry bit is NOT set
const JCN_ZERO_UNSET = 0x1C  // Jump if accumulator is NOT zero

type DecoderFlag struct {
	Name    string
	Value   int
	Changed bool
}

type Decoder struct {
	// Control Flags
	Flags              map[int]DecoderFlag
	DecodedInstruction string // For the renderer
	InstChanged        bool   // For the renderer

	clockCount      int // Internal clock count
	instPhase       int // which instruction phase are we in
	syncSent        bool
	currInstruction int
	dblInstruction  int  // The current instruction requires two complete cycles
	inhibitPCInc    bool // Inhibit the program counter increment for jumps, etc.
	inhibitPC       bool // Block the program counter from writing to the external bus
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
	AccInst                  // Execute an accumulator instruction
	TempLoad                 // Load the temp register from the internal bus
	TempOut                  // Temp register should drive the bus
	AluOut                   // ALU core should drive the bus
	AluEval                  // ALU should evaluate
	AluMode                  // The current mode for the ALU
	ScratchPadIndex          // Which scratchpd (index) register to read/write
	ScratchPadLoad4          // Load 4 bits into the currently selected scratchpad register
	ScratchPadLoad8          // Load 8 bits into the currently selected scratchpad registers
	ScratchPadOut            // Currently selected scratchpad register should drive the bus
	ScratchPadInc            // Currently selected scratchpad register should be incremented
	StackPush                // Push the current address onto the stack
	StackPop                 // Pop the address stack
	DecodeInstruction        // The instruction register is ready to be decoded
	EvalulateJCN             // Evaluate the condition flags for a JCN instruction
	EvalulateISZ             // Evaluate the scratchpad regiser for an ISZ instruction
	END                      // Marker for end of list
)

func (d *Decoder) Init() {
	d.DecodedInstruction = "NOP"
	d.InstChanged = true
	d.clockCount = 0
	d.syncSent = false
	d.currInstruction = -1
	d.Flags = make(map[int]DecoderFlag)
	d.Flags[Sync] = DecoderFlag{"SYNC", 0, false}
	d.Flags[BusDir] = DecoderFlag{"BDIR", 0, false}
	d.Flags[BusTurnAround] = DecoderFlag{"BTA ", 0, false}
	d.Flags[InstRegOut] = DecoderFlag{"INSO  ", 0, false}
	d.Flags[InstRegLoad] = DecoderFlag{"INSL  ", 0, false}
	d.Flags[PCOut] = DecoderFlag{"PCO ", 0, false}
	d.Flags[PCLoad] = DecoderFlag{"PCL ", 0, false}
	d.Flags[PCInc] = DecoderFlag{"PCI ", 0, false}
	d.Flags[AccOut] = DecoderFlag{"ACCO  ", 0, false}
	d.Flags[AccLoad] = DecoderFlag{"ACCL  ", 0, false}
	d.Flags[AccInst] = DecoderFlag{"ACCL  ", -1, false}
	d.Flags[TempOut] = DecoderFlag{"TMPO  ", 0, false}
	d.Flags[TempLoad] = DecoderFlag{"TMPL  ", 0, false}
	d.Flags[AluOut] = DecoderFlag{"ALUO", 0, false}
	d.Flags[AluEval] = DecoderFlag{"ALUE", 0, false}
	d.Flags[AluMode] = DecoderFlag{"ALUM", 0, false}
	d.Flags[ScratchPadIndex] = DecoderFlag{"SPI ", -1, false}
	d.Flags[ScratchPadLoad4] = DecoderFlag{"SPL4", 0, false}
	d.Flags[ScratchPadLoad8] = DecoderFlag{"SPL8", 0, false}
	d.Flags[ScratchPadOut] = DecoderFlag{"SPO ", 0, false}
	d.Flags[ScratchPadInc] = DecoderFlag{"SP+ ", 0, false}
	d.Flags[StackPush] = DecoderFlag{"PUSH", 0, false}
	d.Flags[StackPop] = DecoderFlag{"POP ", 0, false}
	d.Flags[DecodeInstruction] = DecoderFlag{"DEC ", 0, false}
	d.Flags[EvalulateJCN] = DecoderFlag{"EJCN", 0, false}
	d.Flags[EvalulateISZ] = DecoderFlag{"EISZ", 0, false}
}

func (d *Decoder) GetClockCount() int {
	return d.instPhase
}

func (d *Decoder) resetFlags() {
	for i := 0; i < END; i++ {
		if i == ScratchPadIndex || i == AccInst {
			d.clearFlag(i, -1)
		} else {
			d.clearFlag(i, 0)
		}
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

	// Continue to decode instructions after clock 5
	if d.clockCount != 5 && d.currInstruction > 0 {
		d.decodeCurrentInstruction(false)
	}

	switch d.clockCount {
	case 0:
		if !d.inhibitPC {
			// Drive the current address (nybble 0) to the external bus
			d.writeFlag(PCOut, 1)
		}
		d.writeFlag(BusDir, common.DirOut)
	case 1:
		if !d.inhibitPC {
			// Drive the current address (nybble 1) to the external bus
			d.writeFlag(PCOut, 1)
		}
		d.writeFlag(BusDir, common.DirOut)
	case 2:
		// Drive the current address (nybble 2) to the external bus
		d.writeFlag(BusDir, common.DirOut)
		d.writeFlag(PCOut, 1)
	case 4:
		if d.syncSent {
			// Read the OPR from the external bus and write it into the instruction register
			d.writeFlag(BusDir, common.DirIn)
			d.writeFlag(InstRegLoad, 1)
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
		} else {
			d.writeFlag(BusDir, common.DirOut)
		}
		d.writeFlag(Sync, 1)
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
	// Note FIN and JIN share the same upper 4 bits
	case FIN & 0xf0:
		err = d.handleFIN_JIN(fullInst, evalResult)
	case JCN:
		fallthrough
	case JMS:
		fallthrough
	case ISZ:
		fallthrough
	case JUN:
		err = d.handleJCN_JMS_ISZ_JUN(fullInst, evalResult)
	case XCH:
		err = d.handleXCH(fullInst, evalResult)
	case LDM:
		err = d.handleLDM(fullInst, evalResult)
	case LD:
		err = d.handleLD(fullInst, evalResult)
	case INC:
		err = d.handleINC(fullInst, evalResult)
	case FIM:
		fallthrough
	case SRC:
		err = d.handleFIM_SRC(fullInst, evalResult)
	case BBL:
		err = d.handleBBL(fullInst, evalResult)
	case ADD:
		err = d.handleADD(fullInst, evalResult)
	case SUB:
		err = d.handleSUB(fullInst, evalResult)
	// Collectively, all the accumulator instructions
	case ACC:
		err = d.handleACC(fullInst, evalResult)
	}

	// These instructions require decoding the entire 8 bits
	switch fullInst {
	case WRR:
		err = d.handleWRR(fullInst, evalResult)
	}
	return err
}

func (d *Decoder) setDecodedInstruction(inst string) {
	d.DecodedInstruction = inst
	d.InstChanged = true
	rlog.Debugf("--- Decoded instruction is: %s", d.DecodedInstruction)
}
