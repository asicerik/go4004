package cpucore

import (
	"alu"
	"common"
	"css"
	"fmt"
	"image"
	"scratchpad"

	"github.com/tfriedel6/canvas"
)

const BusWidth = 4
const NumRegisters = 16

// Core contains all the components of our cpu
type Core struct {
	regs            scratchpad.Registers
	alu             alu.Alu
	internalDataBus common.Bus
	pc              int
}

// Init create and initialize all the core components
func (c *Core) Init() {
	c.internalDataBus.Init("Internal Data Bus")
	c.regs.Init(&c.internalDataBus.Data, BusWidth, NumRegisters)
	c.alu.Init(&c.internalDataBus.Data, BusWidth)
}

func (c *Core) Step() {
	c.regs.Select(c.pc)
	c.internalDataBus.Data = uint64(c.pc)
	c.regs.Write()
	c.pc++
	if c.pc > 15 {
		c.pc = 0
	}
}

// Test - run a simple test
func (c *Core) Test() {
	// Write each of our registers and make sure we can read them back
	for i := 0; i < 16; i++ {
		c.regs.Select(i)
		c.internalDataBus.Data = uint64(i)
		c.regs.Write()
	}

	// Write the ALU accumulator register
	c.internalDataBus.Data = 0xA
	c.alu.WriteAccumulator()

	for i := 0; i < 16; i++ {
		c.regs.Select(i)
		c.regs.Read()
		fmt.Printf("Register %d is %d\n", i, c.internalDataBus.Data)
	}

	c.alu.ReadAccumulator()
	fmt.Printf("Accumulator is 0x%x\n", c.internalDataBus.Data)
}

func (c *Core) Render(canvas *canvas.Canvas) {
	// Save the current state
	canvas.Save()
	canvas.SetFillStyle("#ccc")
	canvas.FillRect(0, 0, float64(canvas.Width()), float64(canvas.Height()))

	c.internalDataBus.InitRender(image.Point{int(css.Margin), int(css.Margin)},
		image.Point{canvas.Width() - int(css.Margin), int(css.Margin)}, 40)

	canvas.Translate(0, css.Margin)
	c.internalDataBus.Render(canvas)
	c.alu.Render(canvas)
	c.regs.Render(canvas)
	// Restore the state
	canvas.Restore()
}
