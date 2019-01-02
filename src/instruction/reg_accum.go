package instruction

import (
	"alu"
	"fmt"
)

func (d *Decoder) handleXCH(fullInst int, evalResult bool) (err error) {

	// Exchange the accumulator and the scratchpad register
	if d.clockCount == 5 {
		d.setDecodedInstruction(fmt.Sprintf("XCH %X", d.currInstruction&0xf))
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
	return err
}

func (d *Decoder) handleLDM(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 5 {
		d.writeFlag(InstRegOut, 1)
		d.setDecodedInstruction(fmt.Sprintf("LDM %X", d.currInstruction&0xf))
	} else if d.clockCount == 6 {
		d.writeFlag(AccLoad, 1)
		d.currInstruction = -1
	}
	return err
}

func (d *Decoder) handleLD(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 6 {
		d.setDecodedInstruction(fmt.Sprintf("LD  %X", d.currInstruction&0xf))
		// Load the data from the selected scratchpad register
		d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
		d.writeFlag(ScratchPadOut, 1)
	} else if d.clockCount == 7 {
		d.writeFlag(AccLoad, 1)
		d.currInstruction = -1
	}
	return err
}

func (d *Decoder) handleINC(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 6 {
		d.setDecodedInstruction(fmt.Sprintf("INC %X", d.currInstruction&0xf))
		// Select the scratchpad register
		d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
		// Increment it
		d.writeFlag(ScratchPadInc, 1)
		d.currInstruction = -1
	}
	return err
}

func (d *Decoder) handleADD(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 5 {
		d.setDecodedInstruction(fmt.Sprintf("ADD %X", d.currInstruction&0xf))
		d.writeFlag(AluMode, alu.AluIntModeAdd)
		// Output the data from the selected scratchpad register
		d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
		d.writeFlag(ScratchPadOut, 1)
	} else if d.clockCount == 6 {
		// Load the value into the temp register
		d.writeFlag(TempLoad, 1)
	} else if d.clockCount == 7 {
		// Evaluate the ALU and write the value into the accumulator
		d.writeFlag(AluEval, 1)
		d.writeFlag(AccLoad, 1)
		d.currInstruction = -1
	}
	return err
}

func (d *Decoder) handleSUB(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 5 {
		d.setDecodedInstruction(fmt.Sprintf("SUB %X", d.currInstruction&0xf))
		d.writeFlag(AluMode, alu.AluIntModeSub)
		// Output the data from the selected scratchpad register
		d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
		d.writeFlag(ScratchPadOut, 1)
	} else if d.clockCount == 6 {
		// Load the value into the temp register
		d.writeFlag(TempLoad, 1)
	} else if d.clockCount == 7 {
		// Evaluate the ALU and write the value into the accumulator
		d.writeFlag(AluEval, 1)
		d.writeFlag(AccLoad, 1)
		d.currInstruction = -1
	}
	return err
}
