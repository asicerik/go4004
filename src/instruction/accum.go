package instruction

import "fmt"

func (d *Decoder) handleACC(fullInst int, evalResult bool) (err error) {
	if d.clockCount == 5 {
		d.setDecodedInstruction(fmt.Sprintf("ACC %X", d.currInstruction&0xf))
		// Output the data from the selected scratchpad register
		d.writeFlag(AccInst, int(d.currInstruction&0xf))
		d.currInstruction = -1
	}
	return err
}
