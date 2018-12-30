package common

// Bus is used to transfer data between elements and also handle graphics rendering
type Bus struct {
	Name     string
	data     uint64
	mask     uint64
	BusWidth int
	writes   int // Number of writes to the bus during this tick
}

func (b *Bus) Init(busWidth int, name string) {
	b.Name = name
	b.BusWidth = busWidth
	for i := 0; i < busWidth; i++ {
		b.mask = b.mask << 1
		b.mask = b.mask | 1
	}

	b.data = 0xffffffffffffffff & b.mask
}

func (b *Bus) Write(value uint64) {
	b.data = value
	b.writes++
}

func (b *Bus) Read() (value uint64) {
	return b.data
}

func (b *Bus) Reset() {
	b.data = 0xffffffffffffffff & b.mask
	b.writes = 0
}
