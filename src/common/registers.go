package common

type Register struct {
	reg     uint8
	dataBus *uint8
}

func (r *Register) Init(dataBus *uint8) {
	r.dataBus = dataBus
}

func (r *Register) Read() {
	value := r.reg & 0x0f
	*r.dataBus = value
}

func (r *Register) Write() {
	r.reg = *r.dataBus
}
