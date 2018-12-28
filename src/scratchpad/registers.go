package scratchpad

type Registers struct {
	// interfaces.BusDriver
	// interfaces.ClockedElement
	regs    []uint8
	index   int
	dataBus *uint8
}

func (r *Registers) Init(dataBus *uint8) {
	r.regs = make([]uint8, 16)
	r.index = 0
	r.dataBus = dataBus
}

func (r *Registers) Read() {
	value := r.regs[r.index] & 0x0f
	*r.dataBus = value
}

func (r *Registers) Select(index int) {
	r.index = index
}

func (r *Registers) Write() {
	r.regs[r.index] = *r.dataBus
}
