package cpucore

import (
	"alu"
	"common"
	"scratchpad"
)

const BusWidth = 4
const NumRegisters = 16

// Core contains all the logic components of our cpu
type Core struct {
	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	pc              int
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.internalDataBus.Init("Internal Data Bus")
	c.regs.Init(&c.internalDataBus, BusWidth, NumRegisters)
	c.alu.Init(&c.internalDataBus, BusWidth)
}

func (c *Core) Step() {
	c.regs.Select(c.pc)
	c.internalDataBus.Write(uint64(c.pc))
	c.regs.Write()
	// c.regs.Log()
	c.pc++
	if c.pc > 15 {
		c.pc = 0
	}
}
