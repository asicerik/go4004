package rom4001

import (
	"common"
	"interfaces"
)

const BusWidth = 4
const DataWidth = 8
const Depth = 256

// Rom4001 is a model of the Intel 4001 ROM
type Rom4001 struct {
	interfaces.ClockedElement
	data           []uint8           // Data array
	chipID         int               // Hard-coded in metal Rom ID
	busExt         *common.Bus       // External Data Bus for address/data
	busInt         common.Bus        // Internal Data Bus for address/data
	ioBuf          common.Buffer     // Bus i/o buffer
	cm             *int              // CM-ROM select from CPU
	sync           *int              // SYNC signal from CPU
	syncLatched    int               // SYNC latched with clock
	syncSeen       bool              // Have we seen the sync flag?
	clockCount     int               // Internal counter for clock timing
	instPhase      int               // Instruction phase
	addressReg     common.Register   // Address shift register
	outputReg      common.Register   // Output data register
	chipSelected   bool              // We are targeted for this read
	valueRegisters []common.Register // The values near the current address
}

func (r *Rom4001) Init(busExt *common.Bus, sync *int) {
	r.busExt = busExt
	r.sync = sync
	r.busInt.Init(BusWidth, "")
	r.addressReg.Init(nil, 12, "Addr = ")
	r.outputReg.Init(&r.busInt, 4, "")
	r.ioBuf.Init(&r.busInt, r.busExt, BusWidth, "I/O BUF")
	r.data = make([]uint8, Depth)
	r.valueRegisters = make([]common.Register, 3)
	for i := range r.valueRegisters {
		r.valueRegisters[i].Init(nil, 8, "      ")
	}

	// Load a sample program into memory
	r.data[0] = 0xD5 // LDM 5
	r.data[1] = 0xB2 // XCH r2
	r.data[2] = 0xDE // LDM 0xE
	r.data[3] = 0xB3 // XCH r3
	r.data[4] = 0x40 // JUN 0
	r.data[5] = 0x00 // JUN 0 (cont)
	// Set the rest to incrementing values
	for i := 6; i < len(r.data); i++ {
		r.data[i] = uint8(i)
	}
	r.calculateValueRegisters()
	r.chipID = 0
}

func (r *Rom4001) calculateValueRegisters() {
	curr := r.addressReg.ReadDirect() & 0xff

	var first uint64
	r.valueRegisters[0].Selected = false
	r.valueRegisters[1].Selected = false
	r.valueRegisters[2].Selected = false

	switch curr {
	case 0:
		first = 0
		r.valueRegisters[0].Selected = r.chipSelected
	case uint64(len(r.data) - 1):
		first = curr - 2
		r.valueRegisters[2].Selected = r.chipSelected
	default:
		first = curr - 1
		r.valueRegisters[1].Selected = r.chipSelected
	}
	r.valueRegisters[0].WriteDirect(uint64(r.data[first]))
	r.valueRegisters[1].WriteDirect(uint64(r.data[first+1]))
	r.valueRegisters[2].WriteDirect(uint64(r.data[first+2]))
}

func (r *Rom4001) GetClockCount() int {
	return r.clockCount
}

func (r *Rom4001) Reset() {
	r.clockCount = 0
}

func (r *Rom4001) Calculate() {
}

// ClockIn clock in external inputs to the core
func (r *Rom4001) ClockIn() {
	r.updateInternal()

}

func (r *Rom4001) updateInternal() {
	r.syncLatched = *r.sync
	r.syncSeen = r.syncSeen || (r.syncLatched == 0)
	if !r.syncSeen {
		return
	}

	switch r.clockCount {
	case 0:
		r.addressReg.WriteDirect(r.addressReg.ReadDirect() | (r.busInt.Read() << (uint(r.clockCount) * 4)))
		// rlog.Debugf("ROM %d: Wrote address register (n0). Curr value=%X", r.chipID, r.addressReg.ReadDirect())
	case 1:
		r.addressReg.WriteDirect(r.addressReg.ReadDirect() | (r.busInt.Read() << (uint(r.clockCount) * 4)))
		// rlog.Debugf("ROM %d: Wrote address register (n1). Curr value=%X", r.chipID, r.addressReg.ReadDirect())
	case 2:
		r.addressReg.WriteDirect(r.addressReg.ReadDirect() | (r.busInt.Read() << (uint(r.clockCount) * 4)))
		// rlog.Debugf("ROM %d: Wrote address register (n2). Curr value=%X", r.chipID, r.addressReg.ReadDirect())
	}
}

// ClockOut clock external outputs to their respective busses/logic lines
func (r *Rom4001) ClockOut() {
	r.busInt.Reset()
	r.instPhase = r.clockCount

	// Defaults
	r.ioBuf.Disable()

	if !r.syncSeen {
		return
	}

	switch r.clockCount {
	case 7:
		fallthrough
	case 0:
		fallthrough
	case 1:
		// Copy from the external bus to the internal bus
		r.ioBuf.BtoA()
	case 2:
		// Reset the external bus since this is a turn-around cycle
		// r.busExt.Reset()
		fallthrough
	case 3:
		romID := (r.addressReg.ReadDirect() >> 8) & 0xf
		addr := r.addressReg.ReadDirect() & 0xff

		r.chipSelected = romID == uint64(r.chipID)
		r.calculateValueRegisters()
		if r.chipSelected {
			// Write to the external bus from the internal bus
			data := uint64(r.data[addr])
			if r.clockCount == 2 {
				data = data >> 4
			}
			// rlog.Debugf("ROM %d: writing %X to bus", r.chipID, data&0xf)
			r.busInt.Write(data & 0xf)
			r.ioBuf.AtoB()
		}
	}
	if r.syncLatched == 0 {
		r.clockCount = 0
		r.addressReg.WriteDirect(0)
	} else {
		if r.clockCount < 7 {
			r.clockCount++
		} else {
			r.clockCount = 0
		}
	}
}
