package cpucore

import (
	"alu"
	"common"
	"math/rand"
	"scratchpad"
)

const BusWidth = 4
const NumRegisters = 16

// Core contains all the logic components of our cpu
type Core struct {
	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	externalDataBus common.Bus
	busBuffer       ExternalBusBuffer
	pc              int
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.internalDataBus.Init("Internal Data Bus")
	c.externalDataBus.Init("External Data Bus")
	c.busBuffer.Init(&c.externalDataBus, &c.internalDataBus, "Bus Buffer")
	c.regs.Init(&c.internalDataBus, BusWidth, NumRegisters)
	c.alu.Init(&c.internalDataBus, BusWidth)
}

func (c *Core) Step() {
	c.internalDataBus.Reset()
	c.regs.Select(c.pc)
	c.internalDataBus.Write(uint64(c.pc))
	c.regs.Write()
	// c.regs.Log()
	c.pc++
	if c.pc > 15 {
		c.pc = 0
	}
	if rand.Intn(4) == 0 {
		c.busBuffer.buf.BtoA()
	} else {
		c.busBuffer.buf.Disable()
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
