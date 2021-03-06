package common

const DirNone = 0 // disconnected from the bus
const DirAtoB = 1 // transferring from A to B
const DirBtoA = 2 // transferring from B to A
const DirOut = 3  // transferring in to out
const DirIn = 4   // transferring out to in

// Buffer represents a bi-directional non-latching bus buffer
type Buffer struct {
	Name     string
	dataBusA *Bus
	dataBusB *Bus
	Dir      int
	width    int
	mask     uint64
	changed  bool
}

func (b *Buffer) Init(busA *Bus, busB *Bus, width int, name string) {
	b.dataBusA = busA
	b.dataBusB = busB
	b.width = width
	for i := 0; i < width; i++ {
		b.mask = b.mask << 1
		b.mask = b.mask | 1
	}
	b.Name = name
	b.Dir = DirBtoA
}

// Disable disconnects the buffer from the bus
func (b *Buffer) Disable() {
	b.Dir = DirNone
	b.changed = true
}

// AtoB transfers data from bus A to bus B
func (b *Buffer) AtoB() {
	value := (*b.dataBusA).Read() & b.mask
	(*b.dataBusB).Write(value)
	b.Dir = DirAtoB
	b.changed = true
}

// BtoA transfers data from bus B to bus A
func (b *Buffer) BtoA() {
	value := (*b.dataBusB).Read() & b.mask
	(*b.dataBusA).Write(value)
	b.Dir = DirBtoA
	b.changed = true
}

// SetDirAtoB just sets the direction from bus A to bus B (for the UI)
func (b *Buffer) SetDirAtoB() {
	b.Dir = DirAtoB
	b.changed = true
}

// SetDirBtoA just sets the direction from bus A to bus B (for the UI)
func (b *Buffer) SetDirBtoA() {
	b.Dir = DirBtoA
	b.changed = true
}
