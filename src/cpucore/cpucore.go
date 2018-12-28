package cpucore

import (
	"alu"
	"fmt"
	"scratchpad"
)

// Core contains all the components of our cpu
type Core struct {
	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus uint8
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.regs.Init(&c.internalDataBus)
	c.alu.Init(&c.internalDataBus)
}

// Test - run a simple test
func (c *Core) Test() {
	// Write each of our registers and make sure we can read them back
	for i := 0; i < 16; i++ {
		c.regs.Select(i)
		c.internalDataBus = uint8(i)
		c.regs.Write()
	}

	// Write the ALU accumulator register
	c.internalDataBus = 0xA
	c.alu.WriteAccumulator()

	for i := 0; i < 16; i++ {
		c.regs.Select(i)
		c.regs.Read()
		fmt.Printf("Register %d is %d\n", i, c.internalDataBus)
	}

	c.alu.ReadAccumulator()
	fmt.Printf("Accumulator is 0x%x\n", c.internalDataBus)
}
