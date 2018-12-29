package common

type Register struct {
	Name    string
	Reg     uint64
	dataBus *Bus
	width   int
	mask    uint64
	changed bool
}

func (r *Register) Init(dataBus *Bus, width int, name string) {
	r.dataBus = dataBus
	r.width = width
	for i := 0; i < width; i++ {
		r.mask = r.mask << 1
		r.mask = r.mask | 1
	}
	r.Name = name
	r.changed = true
}

func (r *Register) Read() {
	value := r.Reg & r.mask
	(*r.dataBus).Write(value)
}

func (r *Register) Write() {
	r.Reg = (*r.dataBus).Read() & r.mask
	r.changed = true
}

// WriteDirect directly writes the register instead of using the bus
func (r *Register) WriteDirect(value uint64) {
	r.Reg = value & r.mask
	r.changed = true
}

// Increment directly increments the value in a register
// Not sure all hardware supports this
func (r *Register) Increment() {
	r.Reg = (r.Reg + 1) & r.mask
	r.changed = true
}
