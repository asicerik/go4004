package addressstack

import (
	"common"
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
func (s *AddressStack) WriteProgramCounter(nybble uint64, in uint64) {
	var mask uint64
	mask = 0xf << (nybble * 4)
	value := ((s.pc.Reg & ^mask) | (in << (nybble * 4) & mask)) & s.mask
	s.pc.WriteDirect(value)
}

// IncProgramCounter increments the program counter
func (s *AddressStack) IncProgramCounter() {
	s.pc.Increment()
}
