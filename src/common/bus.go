package common

// Bus is used to transfer data between elements and also handle graphics rendering
type Bus struct {
	Name     string
	data     uint64
	BusWidth int
	writes   int // Number of writes to the bus during this tick
}

func (b *Bus) Init(name string) {
	b.Name = name
	b.data = 0xffffffffffffffff
}

func (b *Bus) Write(value uint64) {
	b.data = value
	b.writes++
}

func (b *Bus) Read() (value uint64) {
	return b.data
}

func (b *Bus) Reset() {
	b.data = 0xffffffffffffffff
	b.writes = 0
}
