package main

import (
	"cpucore"
	"css"
	"fmt"
	"image"
	"instruction"
	"os"
	"rom4001"
	"time"

	"github.com/romana/rlog"

	"github.com/tfriedel6/canvas/glfwcanvas"
)

func main() {

	// Programmatically change an rlog setting from within the program
	os.Setenv("RLOG_LOG_LEVEL", "INFO")
	//os.Setenv("RLOG_TRACE_LEVEL", "0")
	os.Setenv("RLOG_LOG_FILE", "go4004.log")
	rlog.UpdateEnv()

	rlog.Info("Welcome to the go 4004 emulator :)")

	wnd, canvas, err := glfwcanvas.CreateWindow(1280, 1024, "Go 4004")
	if err != nil {
		fmt.Println("Could not open Window")
		return
	}
	canvas.SetFont("C:\\Windows\\Fonts\\courbd.ttf", 24)
	defer wnd.Close()

	core := cpucore.Core{}
	core.Init()
	rom := rom4001.Rom4001{}
	rom.Init(&core.ExternalDataBus, &core.Sync)
	WriteROM(&rom)
	romRenderer := rom4001.Renderer{}
	romLeft := int(css.Margin) + 40

	romRenderer.InitRender(&rom, canvas, image.Rectangle{
		image.Point{romLeft, int(css.Margin)},
		image.Point{romLeft, int(css.Margin)}})
	romHeight := romRenderer.Bounds().Dy()

	coreRenderer := cpucore.Renderer{}
	coreRenderer.InitRender(&core, canvas, image.Rectangle{
		image.Point{int(css.Margin), int(css.Margin) + romHeight},
		image.Point{canvas.Width() - int(2*css.Margin), canvas.Height() - int(2*css.Margin) - romHeight}})

	lastTime := time.Now()
	renderCount := 0
	wnd.MainLoop(func() {
		currTime := time.Now()
		if currTime.Sub(lastTime).Seconds() >= 0.1 {
			lastTime = currTime
			DumpState(core, rom)
			core.Calculate()
			core.ClockIn()
			rom.ClockIn()
			core.ClockOut()
			rom.ClockOut()
			// Render twice because glfw is double buffered
			renderCount = 2
		}
		if renderCount > 0 {
			coreRenderer.Render(canvas)
			romRenderer.Render(canvas)
			canvas.SetFillStyle("#ccc")
			canvas.FillRect(20, float64(canvas.Height())-70, 200, 40)
			canvas.SetFillStyle("#000")
			canvas.FillText(fmt.Sprintf("FPS=%3.1f", wnd.FPS()), 20, float64(canvas.Height())-40)
			renderCount--
		}
	})

	// var loops = 1000000
	// for i := 0; i < loops; i++ {
	// 	// DumpState(core, rom)
	// 	core.Calculate()
	// 	core.ClockIn()
	// 	rom.ClockIn()
	// 	core.ClockOut()
	// 	rom.ClockOut()
	// }
	// duration := time.Now().Sub(lastTime).Seconds()
	// hz := float64(loops) / duration
	// rlog.Infof("Elapsed time = %f seconds, or %3.1f kHz", duration, hz/1000)
	rlog.Info("Goodbye")
}

func DumpState(core cpucore.Core, rom rom4001.Rom4001) {
	rlog.Infof("PC=%X, DBUS=%X, INST=%X, SYNC=%d, CCLK=%d, ROMCLK=%d",
		core.GetProgramCounter(),
		core.ExternalDataBus.Read(),
		core.GetInstructionRegister(),
		core.Sync, core.GetClockCount(),
		rom.GetClockCount())
}

func WriteROM(r *rom4001.Rom4001) {
	data := make([]uint8, rom4001.Depth)
	// Load a sample program into memory
	data[0] = instruction.LDM | 5           // Load 5 into the accumulator
	data[1] = instruction.XCH | 2           // Swap accumulator with r2
	data[2] = instruction.LDM | 0xC         // Load C into the accumulator
	data[3] = instruction.NOP               // NOP
	data[4] = instruction.FIM_SRC | (2) | 1 // Send address in r2,r3 to ROM/RAM
	data[5] = instruction.WRR               // Write accumulator to ROM
	data[6] = 0x40                          // JUN 0
	data[7] = 0x00                          // JUN 0 (cont)
	// Set the rest to incrementing values
	for i := 8; i < len(data); i++ {
		data[i] = uint8(i)
	}
	r.LoadProgram(data)
}
