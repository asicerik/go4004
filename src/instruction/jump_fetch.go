package instruction

import (
	"fmt"

	"github.com/romana/rlog"
)

func (d *Decoder) handleJCN_JMS_ISZ_JUN(fullInst int, evalResult bool) (err error) {
	opr := d.currInstruction & 0xf0

	// Are we on the first phase?
	if d.dblInstruction == 0 {
		if opr == JCN {
			d.writeFlag(EvalulateJCN, 1)
		} else if opr == ISZ {
			d.writeFlag(EvalulateISZ, 1)
		}
		if d.clockCount == 5 {
			if opr == JCN {
				d.setDecodedInstruction(fmt.Sprintf("JCN %X", d.currInstruction&0xf))
			} else if opr == JUN {
				d.setDecodedInstruction(fmt.Sprintf("JUN %X", d.currInstruction&0xf))
			} else if opr == JMS {
				d.setDecodedInstruction(fmt.Sprintf("JMS %X", d.currInstruction&0xf))
			} else if opr == ISZ {
				d.setDecodedInstruction(fmt.Sprintf("ISZ %X", d.currInstruction&0xf))
			}

			d.writeFlag(InstRegOut, 1)
		} else if d.clockCount == 6 {
			if opr == ISZ {
				d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xf))
				d.writeFlag(ScratchPadInc, 1)
			} else {
				// Store the upper 4 bits in the temp register
				d.writeFlag(TempLoad, 1)
			}
			d.dblInstruction = d.currInstruction
			d.currInstruction = -1
		}
	} else {
		if d.clockCount == 5 {
			// If this is a conditional jump, evaluate the condition here
			blockJump := false
			if opr == JCN || opr == ISZ {
				blockJump = !evalResult
			}
			if !blockJump {
				// Block the PC increment
				d.inhibitPCInc = true
				// Output the lower 4 bits of the instruction register
				// It contains the lowest 4 bits of the address
				d.writeFlag(InstRegOut, 1)
			} else {
				rlog.Info("Conditional jump was not taken")
				d.dblInstruction = 0
				d.currInstruction = -1
			}
			if opr == JMS {
				// Push the current address onto the stack
				d.writeFlag(StackPush, 1)
			}
		} else if d.clockCount == 6 {
			// Load the lowest 4 bits into the PC
			d.writeFlag(PCLoad, 1)
			// Output the higher 4 bits of the instruction register
			// It contains the middle 4 bits of the address
			d.writeFlag(InstRegOut, 2)
		} else if d.clockCount == 7 {
			// Load the middle 4 bits into the PC
			d.writeFlag(PCLoad, 2)
			// Output the temp register
			// It contains the middle 4 bits of the address
			d.writeFlag(TempOut, 1)
		} else if d.clockCount == 0 {
			if opr == JUN || opr == JMS {
				// NOTE: we have already started outputting the PC onto the bus
				// for the next cycle, but we can still update the highest bits
				// since they go out last
				// Load the highest 4 bits into the PC
				d.writeFlag(PCLoad, 3)
			}
			d.dblInstruction = 0
			d.currInstruction = -1
			// Unblock the PC increment
			d.inhibitPCInc = false
		}
	}
	return err
}

func (d *Decoder) handleFIN_JIN(fullInst int, evalResult bool) (err error) {
	if (fullInst & 0xf1) == JIN {
		// Jump indirect to address in specified register pair
		if d.clockCount == 6 {
			d.setDecodedInstruction(fmt.Sprintf("JIN %X", d.currInstruction&0xe))
			// Output the lower address to the program counter
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)+0) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadOut, 1)
			// Block the PC increment
			d.inhibitPCInc = true
		} else if d.clockCount == 7 {
			// Load the lowest 4 bits into the PC
			d.writeFlag(PCLoad, 1)

			// Output the lower address to the program counter
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)+1) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadOut, 1)
		} else if d.clockCount == 0 {
			// Load the middle 4 bits into the PC
			d.writeFlag(PCLoad, 2)
			d.currInstruction = -1
			// Unblock the PC increment
			d.inhibitPCInc = false
		}
	} else if (fullInst & 0xf1) == FIN {
		// Fetch indirect to address in register pair 0
		// then store the result in specified register pair
		if d.clockCount == 0 {
			// Output the lower address to the data bus
			d.writeFlag(ScratchPadIndex, 0)
			d.writeFlag(ScratchPadOut, 1)
			// UnBlock the PC increment
			d.inhibitPCInc = false
			// Disable the program counter from using the bus
			d.inhibitPC = true
			// Mark this as a double instruction to prevent the instruction register
			// from being clobbered
			d.dblInstruction = d.currInstruction
		} else if d.clockCount == 1 {
			// Output the middle address to the data bus
			d.writeFlag(ScratchPadIndex, 1)
			d.writeFlag(ScratchPadOut, 1)
		} else if d.clockCount == 2 {
			// Unblock the PC
			d.inhibitPC = false
		} else if d.clockCount == 4 && d.dblInstruction > 0 {
			// Load the ROM data into the scratch pad register pair 0
			d.writeFlag(ScratchPadIndex, int(d.dblInstruction&0xe)+0) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadLoad4, 1)
		} else if d.clockCount == 5 && d.dblInstruction > 0 {
			d.setDecodedInstruction(fmt.Sprintf("FIN %X", d.currInstruction&0xf))
			// Load the ROM data into the scratch pad register pair 1
			d.writeFlag(ScratchPadIndex, int(d.dblInstruction&0xe)+1) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadLoad4, 1)
			// Done
			d.dblInstruction = 1
			d.currInstruction = -1
		} else if d.clockCount == 6 && d.dblInstruction <= 0 {
			// Block the PC increment
			d.inhibitPCInc = true
		}
	}
	return err
}

func (d *Decoder) handleBBL(fullInst int, evalResult bool) (err error) {
	// Branch back and address stack pop
	if d.clockCount == 6 {
		d.setDecodedInstruction(fmt.Sprintf("BBL %X", d.currInstruction&0xf))
		// Pop the address stack
		d.writeFlag(StackPop, 1)
		// Store the data passed into the accumulator
		d.writeFlag(InstRegOut, 1)
		// NOTE : the stack pointer contains the address where the jump was.
		// The incrementer will fire and add 1
	} else if d.clockCount == 7 {
		d.writeFlag(AccLoad, 1)
		d.currInstruction = -1
	}
	return err
}
