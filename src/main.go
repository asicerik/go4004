package main

import (
	"cpucore"
	"css"
	"fmt"
	"image"
	"os"
	"rom4001"
	"time"

	"github.com/romana/rlog"

	"github.com/tfriedel6/canvas/glfwcanvas"
)

func main() {

	// Programmatically change an rlog setting from within the program
	os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
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

	renderCount := 0
	lastTime := time.Now()
	wnd.MainLoop(func() {
		currTime := time.Now()
		if currTime.Sub(lastTime).Seconds() >= 0.1 {
			lastTime = currTime
			core.Step()
			rom.Clock()
			DumpState(core, rom)
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

	rlog.Info("Goodbye")
}

func DumpState(core cpucore.Core, rom rom4001.Rom4001) {
	rlog.Infof("DBUS=%X, INST=%X, SYNC=%d, CCLK=%d, ROMCLK=%d",
		core.ExternalDataBus.Read(),
		core.GetInstructionRegister(),
		core.Sync, core.GetClockCount(),
		rom.GetClockCount())
}
