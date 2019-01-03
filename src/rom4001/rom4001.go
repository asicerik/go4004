package rom4001

import (
	"common"
	"interfaces"
	"supportcommon"
)

const BusWidth = 4
const Depth = 256

// Rom4001 is a model of the Intel 4001 ROM
type Rom4001 struct {
	interfaces.ClockedElement
	Core   supportcommon.RamRom
	busInt common.Bus // Internal Data Bus for address/data
}

func (r *Rom4001) Init(busExt *common.Bus, sync *int, cm *int) {
	r.busInt.Init(BusWidth, "ROM Internal")
	r.Core.Init(busExt, &r.busInt, sync, cm, BusWidth, Depth)
}

func (r *Rom4001) SetIOBus(bus *common.Bus) {
	r.Core.SetIOBus(bus)
}

func (r *Rom4001) LoadProgram(data []uint8) {
	r.Core.LoadProgram(data)
}

func (r *Rom4001) SetChipID(id int) {
	r.Core.SetChipID(id)
}

func (r *Rom4001) GetClockCount() int {
	return r.Core.GetClockCount()
}

func (r *Rom4001) Reset() {
	r.Core.Reset()
}

// ClockIn clock in external inputs to the Core
func (r *Rom4001) ClockIn() {
	r.Core.ClockIn()
}

// ClockOut clock external outputs to their respective busses/logic lines
func (r *Rom4001) ClockOut() {
	r.Core.ClockOut()
}
