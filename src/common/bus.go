package common

// Bus is used to transfer data between elements and also handle graphics rendering
type Bus struct {
	Name     string
	data     uint64
	BusWidth int
	changed  bool
}

func (b *Bus) Init(name string) {
	b.Name = name
}

func (b *Bus) Write(value uint64) {
	b.data = value
	b.changed = true
}

func (b *Bus) Read() (value uint64) {
	return b.data
}
