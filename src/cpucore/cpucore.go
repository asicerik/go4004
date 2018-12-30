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

func (c *Core) Clock() {
	c.internalDataBus.Reset()

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
		c.regs.Select(c.getDecoderFlag(instruction.ScratchPadIndex))
		c.regs.Read()
	}

	if c.getDecoderFlag(instruction.BusDir) == common.DirOut {
		c.busBuffer.buf.BtoA()
	} else if c.getDecoderFlag(instruction.BusDir) == common.DirIn {
		c.busBuffer.buf.AtoB()
	}

	if c.getDecoderFlag(instruction.InstRegLoad) != 0 {
		// Read the OPR from the external bus and write it into the instruction register
		c.inst.Write()
	}

	if c.getDecoderFlag(instruction.DecodeInstruction) != 0 {
		// Write the completed instruction to the decoder
		c.Decoder.SetCurrentInstruction(c.inst.GetInstructionRegister())
	}

	// Special handling for turn-around cycles
	if c.getDecoderFlag(instruction.BusTurnAround) != 0 {
		// To prevent bus collision warning. Ideally, we should make sure write count == 1 first
		c.internalDataBus.Reset()
	}
	if c.getDecoderFlag(instruction.InstRegOut) != 0 {
		c.inst.ReadOPR()
		if c.getDecoderFlag(instruction.BusTurnAround) != 0 {
			// Now turn around and drive the instruction register OPR to the external bus
			c.busBuffer.buf.BtoA()
			rlog.Infof("Driving OPR")
		}
	}

	// Finally, any internal bus loads. Do this last to make sure the bus has valid data
	if c.getDecoderFlag(instruction.AccLoad) != 0 {
		c.alu.WriteAccumulator()
	}
	if c.getDecoderFlag(instruction.TempLoad) != 0 {
		c.alu.WriteTemp()
	}
	if c.getDecoderFlag(instruction.PCLoad) != 0 {
		c.as.WriteProgramCounter(0, 0)
		c.as.WriteProgramCounter(1, 0)
		c.as.WriteProgramCounter(2, 0)
	}
	if c.getDecoderFlag(instruction.ScratchPadLoad4) != 0 {
		c.regs.Select(c.getDecoderFlag(instruction.ScratchPadIndex))
		c.regs.Write()
	}
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
