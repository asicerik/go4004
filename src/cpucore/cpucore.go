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

	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	busBuffer       ExternalBusBuffer
	as              addressstack.AddressStack
	inst            instruction.InstructionReg
	clockCount      int
	syncSent        bool
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
	c.Sync = 0
	c.clockCount = 0 // Start mid cycle so we can catch the next Sync
	c.syncSent = false
}

func (c *Core) GetClockCount() int {
	return c.clockCount
}

func (c *Core) GetInstructionRegister() uint64 {
	return c.inst.GetInstructionRegister()
}

func (c *Core) Step() {
	c.internalDataBus.Reset()

	if c.clockCount < 7 {
		c.clockCount++
	} else {
		c.clockCount = 0
		// Handle startup condition. Make sure we send sync before incrementing
		// the program counter
		if c.syncSent {
			c.as.IncProgramCounter()
		}
		c.syncSent = c.Sync == 0
	}

	// Defaults
	c.Sync = 1
	c.busBuffer.buf.Disable()

	switch c.clockCount {
	case 0:
		// Drive the current address (nybble 0) to the external bus
		c.as.ReadProgramCounter(0)
		c.busBuffer.buf.BtoA()
	case 1:
		// Drive the current address (nybble 1) to the external bus
		c.as.ReadProgramCounter(1)
		c.busBuffer.buf.BtoA()
	case 2:
		// Drive the current address (nybble 2) to the external bus
		c.as.ReadProgramCounter(2)
		c.busBuffer.buf.BtoA()
	case 4:
		fallthrough
	case 5:
		if c.syncSent {
			// Read the OPR from the external bus and write it into the instruction register
			c.busBuffer.buf.AtoB()
			c.inst.Write(c.clockCount - 4)
			if c.clockCount == 5 {
				// Now turn around and drive the instruction register OPR to the external bus
				// To prevent bus collision warning. Ideally, we should make sure write count == 1 first
				c.internalDataBus.Reset()
				c.inst.ReadOPR()
				c.busBuffer.buf.BtoA()
				rlog.Infof("Driving OPR")
			}
		}
	case 7:
		c.Sync = 0
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
