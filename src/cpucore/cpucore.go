package cpucore

import (
	"addressstack"
	"alu"
	"common"
	"instruction"
	"scratchpad"

	"github.com/romana/rlog"
)

const BusWidth = 4
const AddressWidth = 12
const NumRegisters = 16

// Core contains all the logic components of our cpu
type Core struct {
	ExternalDataBus common.Bus
	Sync            int
	Decoder         instruction.Decoder

	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	busBuffer       ExternalBusBuffer
	as              addressstack.AddressStack
	inst            instruction.Instruction
	evaluationFn    func() bool // Conditional jump evaluation function
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.internalDataBus.Init(BusWidth, "Internal Data Bus")
	c.ExternalDataBus.Init(BusWidth, "External Data Bus")
	c.busBuffer.Init(&c.ExternalDataBus, &c.internalDataBus, "Bus Buffer")
	c.regs.Init(&c.internalDataBus, BusWidth, NumRegisters)
	c.alu.Init(&c.internalDataBus, BusWidth)
	c.as.Init(&c.internalDataBus, AddressWidth, 3)
	c.inst.Init(&c.internalDataBus, BusWidth)
	c.Decoder.Init()
	c.Sync = 0
}

func (c *Core) GetClockCount() int {
	return c.Decoder.GetClockCount()
}

func (c *Core) GetInstructionRegister() uint64 {
	return c.inst.GetInstructionRegister()
}

func (c *Core) GetProgramCounter() uint64 {
	return c.as.GetProgramCounter()
}

func (c *Core) getDecoderFlag(index int) int {
	return c.Decoder.Flags[index].Value
}

// Calculate the internal logic before the next clock edge
func (c *Core) Calculate() {
	c.Decoder.CalculateFlags()
}

// ClockIn clock in external inputs to the core
func (c *Core) ClockIn() {
	// Load the data from the external bus if needed
	if c.getDecoderFlag(instruction.BusDir) == common.DirIn {
		c.busBuffer.buf.AtoB()
	}

	c.regs.Select(c.getDecoderFlag(instruction.ScratchPadIndex))

	if c.getDecoderFlag(instruction.InstRegLoad) != 0 {
		// Read the OPR from the external bus and write it into the instruction register
		c.inst.Write()
	}

	if c.getDecoderFlag(instruction.DecodeInstruction) != 0 {
		evalResult := true
		if c.evaluationFn != nil {
			evalResult = c.evaluationFn()
		}
		c.evaluationFn = nil

		// Write the completed instruction to the decoder
		c.Decoder.SetCurrentInstruction(c.inst.GetInstructionRegister(), evalResult)
		c.regs.Select(c.getDecoderFlag(instruction.ScratchPadIndex))
	}

	// Finally, any internal bus loads. Do this last to make sure the bus has valid data
	if c.getDecoderFlag(instruction.AccLoad) != 0 {
		c.alu.WriteAccumulator()
	}
	if c.getDecoderFlag(instruction.TempLoad) != 0 {
		c.alu.WriteTemp()
	}
	if c.getDecoderFlag(instruction.PCLoad) != 0 {
		c.as.WriteProgramCounter(uint64(c.getDecoderFlag(instruction.PCLoad) - 1))
	}
	if c.getDecoderFlag(instruction.ScratchPadLoad4) != 0 {
		c.regs.Write()
	}
}

// ClockOut clock external outputs to their respective busses/logic lines
func (c *Core) ClockOut() {
	c.internalDataBus.Reset()
	c.ExternalDataBus.Reset()

	c.Decoder.Clock()

	// Defaults
	c.busBuffer.buf.Disable()

	if c.getDecoderFlag(instruction.Sync) != 0 {
		c.Sync = 0
		c.inst.Reset()
	} else {
		c.Sync = 1
	}

	if c.getDecoderFlag(instruction.PCInc) != 0 {
		c.as.IncProgramCounter()
	}

	// All actions which output on the data bus need to go before the bus direction selection
	if c.getDecoderFlag(instruction.PCOut) != 0 {
		c.as.ReadProgramCounter(uint64(c.Decoder.GetClockCount()))
	}
	if c.getDecoderFlag(instruction.AccOut) != 0 {
		c.alu.ReadAccumulator()
	}
	if c.getDecoderFlag(instruction.TempOut) != 0 {
		c.alu.ReadTemp()
	}
	if c.getDecoderFlag(instruction.ScratchPadOut) != 0 {
		c.regs.Read()
	}

	if c.getDecoderFlag(instruction.BusDir) == common.DirOut {
		c.busBuffer.buf.BtoA()
	} else if c.getDecoderFlag(instruction.BusDir) == common.DirIn {
		c.busBuffer.buf.SetDirAtoB()
	}

	// // Special handling for turn-around cycles
	// if c.getDecoderFlag(instruction.BusTurnAround) != 0 {
	// 	// To prevent bus collision warning. Ideally, we should make sure write count == 1 first
	// 	c.internalDataBus.Reset()
	// }

	if c.getDecoderFlag(instruction.InstRegOut) != 0 {
		c.inst.ReadInstructionRegister(uint64(c.getDecoderFlag(instruction.InstRegOut)) - 1)
		// if c.getDecoderFlag(instruction.BusTurnAround) != 0 {
		// 	// Now turn around and drive the instruction register OPR to the external bus
		// 	c.inst.ReadOPR()
		// 	c.busBuffer.buf.BtoA()
		// 	rlog.Infof("Driving OPR")
		// }
	}

	// Condtitional evaluation flags
	if c.getDecoderFlag(instruction.EvalulateJCN) != 0 {
		c.evaluationFn = c.evalulateJCN
	}
}

// If these functions return false, conditional jumps are blocked
func (c *Core) evalulateJCN() bool {
	// Not sure how the real CPU does this, so I am cutting corners here
	condititonFlags := c.alu.ReadTempDirect()
	aluFlags := c.alu.ReadFlags()
	// testBitFlag := int(condititonFlags & 0x1)
	carryBitFlag := int((condititonFlags >> 1) & 0x1)
	zeroBitFlag := int((condititonFlags >> 2) & 0x1)
	invertBitFlag := int((condititonFlags >> 3) & 0x1)
	result := true
	if invertBitFlag == 0 {
		result = ((carryBitFlag == 0) || (aluFlags.Carry == 1)) &&
			((zeroBitFlag == 0) || (aluFlags.Zero == 1))
	} else {
		result = ((carryBitFlag == 1) && (aluFlags.Carry == 0)) ||
			((zeroBitFlag == 1) && (aluFlags.Zero == 0))
	}
	rlog.Debugf("evalulateJCN: conditionalFlags=%X, aluFlags=%v. Result=%v", condititonFlags, aluFlags, result)
	return result
}

// ExternalBusBuffer is the buffer that connects the internal and external data busses
type ExternalBusBuffer struct {
	buf    common.Buffer
	busInt *common.Bus
	busExt *common.Bus
}

// Init ...
func (b *ExternalBusBuffer) Init(busExt *common.Bus, busInt *common.Bus, name string) {
	b.buf.Init(busExt, busInt, BusWidth, name)
	b.busExt = busExt
	b.busInt = busInt
}
