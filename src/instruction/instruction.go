package instruction

import (
	"common"

	"github.com/romana/rlog"
)

const InstructionWidth = 16

type Instruction struct {
	busReg     common.Register
	instReg    common.Register
	dataBus    *common.Bus
	width      int
	mask       uint64
	drivingBus bool
	writeCount int
}

func (r *Instruction) Init(dataBus *common.Bus, width int) {
	r.busReg.Init(dataBus, width, "I/O ")
	r.instReg.Init(nil, 8, "INST ")
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}

	r.dataBus = dataBus
}

func (r *Instruction) GetInstructionRegister() uint64 {
	return r.instReg.ReadDirect()
}

// Reset ...
func (r *Instruction) Reset() {
	r.instReg.WriteDirect(0)
	r.drivingBus = false
	r.writeCount = 0
}

func (r *Instruction) ReadInstructionRegister(nybble uint64) {
	value := r.instReg.ReadDirect()
	if nybble == 1 {
		value = value >> 4
	}
	r.busReg.WriteDirect(value)
	r.busReg.Read()
	r.drivingBus = true
}

func (r *Instruction) Write() {
	r.busReg.Write()
	rlog.Tracef(0, "Instruction: Write %X. writeCount=%d", r.busReg.ReadDirect(), r.writeCount)
	if r.writeCount == 0 {
		r.instReg.WriteDirect(r.busReg.ReadDirect() << 4)
	} else {
		tmp := r.instReg.ReadDirect()
		r.instReg.WriteDirect(tmp | r.busReg.ReadDirect())
	}
	r.writeCount++
}
