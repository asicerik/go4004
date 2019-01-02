package main

import (
	"common"
	"cpucore"
	"instruction"
	"os"
	"rom4001"
	"time"

	"github.com/romana/rlog"
)

func main() {

	enableLog := true
	// Programmatically change an rlog setting from within the program
	if enableLog {
		os.Setenv("RLOG_LOG_LEVEL", "ERROR")
		//os.Setenv("RLOG_TRACE_LEVEL", "0")
		os.Setenv("RLOG_LOG_FILE", "go4004.log")
		rlog.UpdateEnv()
	}

	rlog.Info("Welcome to the go 4004 emulator :)")

	core := cpucore.Core{}
	core.Init()

	ioBus := common.Bus{}
	ioBus.Init(4, "ROM I/O bus")
	rom := rom4001.Rom4001{}
	rom.Init(&core.ExternalDataBus, &core.Sync, &core.CmROM)
	rom.SetIOBus(&ioBus)
	WriteROM(&rom)

	lastTime := time.Now()
	var loops = 1000000
	for i := 0; i < loops; i++ {
		if enableLog {
			DumpState(core, rom, &ioBus)
		}
		core.Calculate()
		core.ClockIn()
		rom.ClockIn()
		core.ClockOut()
		rom.ClockOut()
	}
	duration := time.Now().Sub(lastTime).Seconds()
	hz := float64(loops) / duration
	rlog.Errorf("Elapsed time = %f seconds, or %3.1f kHz", duration, hz/1000)
	rlog.Info("Goodbye")
}

func DumpState(core cpucore.Core, rom rom4001.Rom4001, romIoBus *common.Bus) {
	rlog.Infof("PC=%X, DBUS=%X, INST=%X, ROMIO=%X, SYNC=%d, CCLK=%d, ROMCLK=%d",
		core.GetProgramCounter(),
		core.ExternalDataBus.Read(),
		core.GetInstructionRegister(),
		romIoBus.Read(),
		core.Sync, core.GetClockCount(),
		rom.GetClockCount())
}

func WriteROM(r *rom4001.Rom4001) {
	// Load a sample program into memory
	data := instruction.LEDCountUsingAdd()
	r.LoadProgram(data)
}
