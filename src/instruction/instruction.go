package instruction

import "common"

type InstructionReg struct {
	reg     common.Register
	dataBus *common.Bus
	width   int
	mask    uint64
}

func (r *InstructionReg) Init(dataBus *common.Bus, width int) {
	r.reg.Init(dataBus, width, "INST ")
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}

	r.dataBus = dataBus
}

func (r *InstructionReg) GetInstructionRegister() uint64 {
	return r.reg.Reg
}

func (r *InstructionReg) ReadOPR() {
	r.reg.Read()
}

func (r *InstructionReg) Write() {
	r.reg.Write()
}
