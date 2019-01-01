package addressstack

import (
	"common"

	"github.com/romana/rlog"
)

// AddressStack contains the program counter, stack pointer and stack address registers
type AddressStack struct {
	pc           common.Register
	stack        []common.Register
	stackPointer int
	dataBus      *common.Bus
	width        int
	mask         uint64
	drivingBus   bool
}

func (s *AddressStack) Init(dataBus *common.Bus, width int, depth int) {
	s.pc.Init(nil, width, "PC    ")
	s.stack = make([]common.Register, depth)
	for i := 0; i < depth; i++ {
		s.stack[i].Init(nil, width, "Level "+string(1))
	}
	s.width = width
	for i := 0; i < width; i++ {
		s.mask = s.mask << 1
		s.mask = s.mask | 1
	}

	s.stackPointer = 0
	s.dataBus = dataBus
	s.pc.WriteDirect(0)
}

// GetProgramCounter is for debugging
func (s *AddressStack) GetProgramCounter() uint64 {
	return s.pc.Reg
}

// ReadProgramCounter reads the program counter one nybble at a time
func (s *AddressStack) ReadProgramCounter(nybble uint64) {
	value := s.pc.Reg >> (nybble * 4) & 0xf
	s.dataBus.Write(value)
	s.drivingBus = true
}

// WriteProgramCounter writes the program counter one nybble at a time
func (s *AddressStack) WriteProgramCounterDirect(nybble uint64, in uint64) {
	var mask uint64
	mask = 0xf << (nybble * 4)
	value := ((s.pc.Reg & ^mask) | (in << (nybble * 4) & mask)) & s.mask
	s.pc.WriteDirect(value)
	rlog.Debugf("AddressStack: Direct Wrote program counter nybble %d. New value=%03X", nybble, value)
}

// WriteProgramCounter writes the program counter one nybble at a time
func (s *AddressStack) WriteProgramCounter(nybble uint64) {
	busValue := s.dataBus.Read()
	var mask uint64
	mask = 0xf << (nybble * 4)
	value := ((s.pc.Reg & ^mask) | (busValue << (nybble * 4) & mask)) & s.mask
	s.pc.WriteDirect(value)
	rlog.Debugf("AddressStack: Wrote program counter nybble %d. New value=%03X", nybble, value)
}

// IncProgramCounter increments the program counter
func (s *AddressStack) IncProgramCounter() {
	s.pc.Increment()
}

func (s *AddressStack) StackPush() {
	if s.stackPointer == len(s.stack) {
		rlog.Warn("Stack overflow")
		return
	}
	rlog.Infof("Stack PUSH: SP=%d (pre), PC=%03X", s.stackPointer, s.pc.Reg)
	s.stack[s.stackPointer].WriteDirect(s.pc.Reg)
	s.stackPointer++
}

func (s *AddressStack) StackPop() {
	if s.stackPointer == 0 {
		rlog.Warn("Stack underflow")
		return
	}
	s.pc.WriteDirect(s.stack[s.stackPointer].ReadDirect())
	rlog.Infof("Stack POP: SP=%d (pre), PC=%03X", s.stackPointer, s.pc.Reg)
	s.stackPointer--
}
