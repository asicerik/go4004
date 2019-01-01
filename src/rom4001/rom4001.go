package rom4001

import (
	"common"
	"instruction"
	"interfaces"

	"github.com/romana/rlog"
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
	busBuf         common.Buffer     // Bus i/o buffer
	cm             *int              // CM-ROM select from CPU
	sync           *int              // SYNC signal from CPU
	syncLatched    int               // SYNC latched with clock
	syncSeen       bool              // Have we seen the sync flag?
	clockCount     int               // Internal counter for clock timing
	instPhase      int               // Instruction phase
	addressReg     common.Register   // Address shift register
	instReg        common.Register   // Instruction shift register
	outputReg      common.Register   // Output data register
	chipSelected   bool              // We are targeted for this read
	valueRegisters []common.Register // The values near the current address
	ioBus          *common.Bus       // Input/Output bus for general purpose IO
	srcDetected    bool              // SRC command was detected
	srcRomID       uint64            // The ROM ID sent in the SRC command
	ioOpDetected   bool              // IO Operation was detected
}

func (r *Rom4001) Init(busExt *common.Bus, sync *int, cm *int) {
	r.busExt = busExt
	r.sync = sync
	r.cm = cm
	r.busInt.Init(BusWidth, "ROM Internal")
	r.addressReg.Init(nil, 12, "Addr = ")
	r.instReg.Init(nil, 8, "Inst = ")
	r.outputReg.Init(&r.busInt, 4, "")
	r.busBuf.Init(&r.busInt, r.busExt, BusWidth, "I/O BUF")
	r.data = make([]uint8, Depth)
	r.valueRegisters = make([]common.Register, 3)
	for i := range r.valueRegisters {
		r.valueRegisters[i].Init(nil, 8, "      ")
	}

	r.calculateValueRegisters()
	r.chipID = 0
}

func (r *Rom4001) SetIOBus(bus *common.Bus) {
	r.ioBus = bus
}

func (r *Rom4001) LoadProgram(data []uint8) {
	copy(r.data, data)
	r.calculateValueRegisters()
}

func (r *Rom4001) SetChipID(id int) {
	r.chipID = id
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
		// Copy from the external bus to the internal bus
		r.busBuf.BtoA()
		r.addressReg.WriteDirect((r.busInt.Read() << (uint(r.clockCount) * 4)))
		rlog.Tracef(0, "ROM %d: Wrote address register (n0). Curr value=%03X", r.chipID, r.addressReg.ReadDirect())
	case 1:
		// Copy from the external bus to the internal bus
		r.busBuf.BtoA()
		r.addressReg.WriteDirect(r.addressReg.ReadDirect() | (r.busInt.Read() << (uint(r.clockCount) * 4)))
		rlog.Tracef(0, "ROM %d: Wrote address register (n1). Curr value=%03X", r.chipID, r.addressReg.ReadDirect())
	case 2:
		// Copy from the external bus to the internal bus
		r.busBuf.BtoA()
		r.addressReg.WriteDirect(r.addressReg.ReadDirect() | (r.busInt.Read() << (uint(r.clockCount) * 4)))
		rlog.Tracef(0, "ROM %d: Wrote address register (n2). Curr value=%03X", r.chipID, r.addressReg.ReadDirect())
		romID := (r.addressReg.ReadDirect() >> 8) & 0xf
		r.chipSelected = (romID == uint64(r.chipID)) && (*(r.cm) == 0)
		if r.chipSelected {
			rlog.Tracef(0, "ROM %d: Selected for read access", r.chipID)
		}
	case 3:
		// Copy from the external bus to the internal bus
		// if we are not writing the bus
		if !r.chipSelected {
			r.busBuf.BtoA()
		}
		// Check for IO ops before we update the SRC flag
		if r.busInt.Read() == (instruction.WRR >> 4) {
			if r.srcDetected {
				rlog.Debug("ROM: IO instruction detected")
				r.ioOpDetected = true
			} else {
				r.ioOpDetected = false
			}
		} else {
			r.ioOpDetected = false
		}

		// NOTE: FIM and SRC have the same upper 4 bits
		// We won't know which instruction it is until the next cycle
		if r.busInt.Read() == (instruction.SRC >> 4) {
			rlog.Debug("ROM: FIM/SRC instruction detected")
			r.srcDetected = true
		} else {
			r.srcDetected = false
		}
	case 4:
		// Copy from the external bus to the internal bus
		// if we are not writing the bus
		if !r.chipSelected {
			r.busBuf.BtoA()
		}
		if r.srcDetected {
			if (r.busInt.Read() & 0x1) == 0 {
				rlog.Debug("ROM: FIM instruction verified")
			} else {
				rlog.Debug("ROM: SRC instruction verified")
			}
		}
		r.instReg.WriteDirect(r.busInt.Read())
	case 6:
		if r.srcDetected {
			// Copy the data to the inernal bus
			r.busBuf.BtoA()
			r.srcRomID = r.busInt.Read() & 0xf
			if int(r.srcRomID) != r.chipID {
				rlog.Debugf("ROM: SRC command was NOT for us. Our chipID=%02X, cmd chipID=%02X",
					r.chipID, r.srcRomID)
				r.srcDetected = false
			} else {
				rlog.Debugf("ROM: SRC command WAS for us. Our chipID=%02X",
					r.chipID)
			}
		}
		if r.ioOpDetected {
			// Copy the data to the inernal bus
			r.busBuf.BtoA()
			cmd := r.instReg.ReadDirect()
			switch cmd {
			case instruction.WRR & 0xf:
				// IO Write
				r.ioBus.Reset()
				r.ioBus.Write(r.busInt.Read())
			}
		}
	}
}

// ClockOut clock external outputs to their respective busses/logic lines
func (r *Rom4001) ClockOut() {
	r.busInt.Reset()
	r.instPhase = r.clockCount

	// Defaults
	r.busBuf.Disable()

	if !r.syncSeen {
		return
	}

	switch r.clockCount {
	case 7:
		fallthrough
	case 0:
		fallthrough
	case 1:
		// Set the direction from the external bus to the internal bus
		// The actual copy happens in ClockIn()
		r.busBuf.SetDirBtoA()
	case 2:
		// Reset the external bus since this is a turn-around cycle
		// r.busExt.Reset()
		fallthrough
	case 3:
		addr := r.addressReg.ReadDirect() & 0xff

		r.calculateValueRegisters()
		if r.chipSelected {
			// Write to the external bus from the internal bus
			data := uint64(r.data[addr])
			if r.clockCount == 2 {
				data = data >> 4
			}
			// rlog.Debugf("ROM %d: writing %X to bus", r.chipID, data&0xf)
			r.busInt.Write(data & 0xf)
			r.busBuf.AtoB()
		} else {
			// Copy from the external bus to the internal bus
			// r.busBuf.BtoA()
		}
	case 5:
		if r.ioOpDetected {
			cmd := r.instReg.ReadDirect()
			switch cmd {
			case instruction.RDR & 0xf:
				// IO Read
				r.busInt.Write(r.ioBus.Read() & 0xf)
				r.busBuf.AtoB()
			}
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
