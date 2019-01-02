package scratchpad

import (
	"common"
	"fmt"

	"github.com/romana/rlog"
)

type Registers struct {
	regs       []common.Register
	index      int
	dataBus    *common.Bus
	width      int
	mask       uint64
	drivingBus bool
}

func (r *Registers) Init(dataBus *common.Bus, width int, depth int) {
	r.regs = make([]common.Register, depth)
	for i := 0; i < depth; i++ {
		r.regs[i].Init(dataBus, width, "")
	}
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}

	r.index = 0
	r.dataBus = dataBus
}

func (r *Registers) Read() {
	r.regs[r.index].Read()
	r.drivingBus = true
}

func (r *Registers) Select(index int) {
	if r.index != index {
		rlog.Debugf("Selected ScratchPad Register %d", index)
	}
	r.index = index
}

func (r *Registers) Write() {
	rlog.Debugf("Writing ScratchPad Register %d with %X", r.index, r.dataBus.Read())
	r.regs[r.index].Write()
}

func (r *Registers) Inc() {
	value := r.regs[r.index].ReadDirect()
	value = (value + 1) & 0xf
	rlog.Debugf("Incremented ScratchPad Register %d. New value is %X", r.index, value)
	r.regs[r.index].WriteDirect(value)
}

func (r *Registers) IsCurrentRegisterZero() bool {
	return r.regs[r.index].ReadDirect() == 0
}

func (r *Registers) Log() {
	ret := ""
	for i := range r.regs {
		ret = ret + fmt.Sprintf("%X ", r.regs[i].Reg)
	}
	rlog.Debug(ret)
}
