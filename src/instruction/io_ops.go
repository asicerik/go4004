package instruction

import "fmt"

func (d *Decoder) handleFIM_SRC(fullInst int, evalResult bool) (err error) {
	if (fullInst & 0x1) == 0 {
		d.setDecodedInstruction("FIM")
	} else {
		// Send I/O address to ROM/RAM
		d.setDecodedInstruction(fmt.Sprintf("SRC %X", (d.currInstruction&0xf)>>1))
		if d.clockCount == 6 {
			// Output the selected scratchpad register
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadOut, 1)
		} else if d.clockCount == 7 {
			// Output the selected scratchpad register + 1
			d.writeFlag(ScratchPadIndex, int(d.currInstruction&0xe)+1) // Note - we are chopping bit 0
			d.writeFlag(ScratchPadOut, 1)
			d.currInstruction = -1
		}
	}
	return err
}

func (d *Decoder) handleWRR(fullInst int, evalResult bool) (err error) {
	// ROM I/O Write
	// Send I/O address to ROM/RAM
	d.setDecodedInstruction("WRR")
	if d.clockCount == 6 {
		// Output the accumulator to the external bus
		d.writeFlag(AccOut, 1)
		d.currInstruction = -1
	}
	return err
}
