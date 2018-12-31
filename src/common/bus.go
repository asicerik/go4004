package common

import "github.com/romana/rlog"

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
	rlog.Debugf("BUS: %s write=%X, writesPre=%d. this=%p", b.Name, value, b.writes, b)
	b.data = value
	b.writes++
	if b.writes > 1 {
		rlog.Warnf("**** Bus collision. Name=%s, writes=%d", b.Name, b.writes)
	}
}

func (b *Bus) Read() (value uint64) {
	return b.data
}

func (b *Bus) Reset() {
	rlog.Tracef(1, "BUS: %s Reset", b.Name)
	b.data = 0xffffffffffffffff & b.mask
	b.writes = 0
}
