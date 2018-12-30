package instruction

import "common"

const InstructionWidth = 16

type InstructionReg struct {
	busReg     common.Register
	instReg    common.Register
	dataBus    *common.Bus
	width      int
	mask       uint64
	drivingBus bool
}

func (r *InstructionReg) Init(dataBus *common.Bus, width int) {
	r.busReg.Init(dataBus, width, "I/O ")
	r.instReg.Init(nil, 8, "INST ")
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}

	r.dataBus = dataBus
}

func (r *InstructionReg) GetInstructionRegister() uint64 {
	return r.instReg.ReadDirect()
}

func (r *InstructionReg) ReadOPR() {
	r.busReg.Read()
	r.drivingBus = true
}

func (r *InstructionReg) Write(nybble int) {
	r.busReg.Write()
	if nybble == 0 {
		r.instReg.WriteDirect(r.busReg.ReadDirect() << 4)
	} else {
		tmp := r.instReg.ReadDirect()
		r.instReg.WriteDirect(tmp | r.busReg.ReadDirect())
	}
}
