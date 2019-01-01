package main

import (
	"common"
	"cpucore"
	"css"
	"fmt"
	"image"
	"instruction"
	"rom4001"
	"time"

	"github.com/romana/rlog"

	"github.com/tfriedel6/canvas/glfwcanvas"
)

func main() {

	// Programmatically change an rlog setting from within the program
	// os.Setenv("RLOG_LOG_LEVEL", "INFO")
	// os.Setenv("RLOG_TRACE_LEVEL", "0")
	// os.Setenv("RLOG_LOG_FILE", "go4004.log")
	// rlog.UpdateEnv()

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

	ioBus := common.Bus{}
	ioBus.Init(4, "ROM I/O bus")
	rom := rom4001.Rom4001{}
	rom.Init(&core.ExternalDataBus, &core.Sync, &core.CmROM)
	rom.SetIOBus(&ioBus)
	WriteROM(&rom)

	romRenderer := rom4001.Renderer{}
	romLeft := int(css.Margin) + 40
	romRenderer.InitRender(&rom, canvas, image.Rectangle{
		image.Point{romLeft, int(css.Margin)},
		image.Point{romLeft, int(css.Margin)}})
	romHeight := romRenderer.Bounds().Dy()
	romWidth := romRenderer.Bounds().Dx()

	led0Renderer := IoBusRenderer{}
	led0Left := romLeft + romWidth + 20
	ledWidth := 120
	ledHeight := 120
	led0Renderer.InitRender(&ioBus, 0, image.Rectangle{
		image.Point{led0Left, int(css.Margin)},
		image.Point{led0Left + ledWidth, int(css.Margin) + ledHeight}})

	coreRenderer := cpucore.Renderer{}
	coreRenderer.InitRender(&core, canvas, image.Rectangle{
		image.Point{int(css.Margin), int(css.Margin) + romHeight},
		image.Point{canvas.Width() - int(2*css.Margin), canvas.Height() - int(2*css.Margin) - romHeight}})

	lastTime := time.Now()
	renderCount := 0
	clocksPerRender := 1
	wnd.MainLoop(func() {
		currTime := time.Now()
		if currTime.Sub(lastTime).Seconds() >= 0.01 {
			lastTime = currTime
			for i := 0; i < clocksPerRender; i++ {
				//DumpState(core, rom, &ioBus)
				core.Calculate()
				core.ClockIn()
				rom.ClockIn()
				core.ClockOut()
				rom.ClockOut()
			}
			// Render twice because glfw is double buffered
			renderCount = 2
		}
		if renderCount > 0 {
			coreRenderer.Render(canvas)
			romRenderer.Render(canvas)
			led0Renderer.Render(canvas)
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
	data := instruction.LEDCount()
	r.LoadProgram(data)
}
