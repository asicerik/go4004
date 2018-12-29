package cpucore

import (
	"addressstack"
	"alu"
	"common"
	"scratchpad"
)

const BusWidth = 4
const AddressWidth = 12
const NumRegisters = 16

// Core contains all the logic components of our cpu
type Core struct {
	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	externalDataBus common.Bus
	busBuffer       ExternalBusBuffer
	as              addressstack.AddressStack
	clockCycle      int
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.internalDataBus.Init("Internal Data Bus")
	c.externalDataBus.Init("External Data Bus")
	c.busBuffer.Init(&c.externalDataBus, &c.internalDataBus, "Bus Buffer")
	c.regs.Init(&c.internalDataBus, BusWidth, NumRegisters)
	c.alu.Init(&c.internalDataBus, BusWidth)
	c.as.Init(&c.internalDataBus, AddressWidth, 3)
}

func (c *Core) Step() {
	c.internalDataBus.Reset()

	switch c.clockCycle {
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
	default:
		c.busBuffer.buf.Disable()
	}
	if c.clockCycle < 8 {
		c.clockCycle++
	} else {
		c.clockCycle = 0
		c.as.IncProgramCounter()
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
