package main

import (
	"common"
	"cpucore"
	"css"
	"fmt"
	"image"
	"instruction"
	"os"
	"rom4001"

	"github.com/romana/rlog"

	"github.com/tfriedel6/canvas/glfwcanvas"
)

type runFlags struct {
	StepClock bool // Step one clock
	StepCycle bool // Step 8 clocks
	FreeRun   bool // Let 'er rip!
	Halt      bool // Stop the processor
	Quit      bool // Quit the program
}

var currentRunFlags = runFlags{}

func KeyDown(scancode int, rn rune, name string) {
	currentRunFlags = runFlags{}
	switch name {
	case "KeyC":
		currentRunFlags.StepClock = true
	case "KeyS":
		currentRunFlags.StepCycle = true
	case "KeyR":
		currentRunFlags.FreeRun = true
	case "Escape":
		fallthrough
	case "KeyQ":
		currentRunFlags.Quit = true
	}
}

func main() {

	enableLog := true
	// Programmatically change an rlog setting from within the program
	if enableLog {
		os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
		os.Setenv("RLOG_TRACE_LEVEL", "0")
		os.Setenv("RLOG_LOG_FILE", "go4004.log")
		rlog.UpdateEnv()
	}

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

	// Set window callbacks
	wnd.KeyDown = KeyDown

	renderCount := 2
	// How many clock cycles do we run between renders
	// 1 = render every clock cycle (= SLOW)
	// Current max performance is about 160kHz on my machine,
	// which works out to about 5,300 clocks max before the frame
	// rate drops. 8192 gives about 20fps on my machine
	clocksPerRender := 1
	cycleCount := 0
	wnd.MainLoop(func() {
		if currentRunFlags.Quit {
			wnd.Close()
		}
		if currentRunFlags.StepClock || currentRunFlags.StepCycle || currentRunFlags.FreeRun {
			for i := 0; i < clocksPerRender; i++ {
				if enableLog {
					DumpState(core, rom, &ioBus)
				}
				core.Calculate()
				core.ClockIn()
				rom.ClockIn()
				core.ClockOut()
				rom.ClockOut()
				cycleCount++
			}
			// Render twice because glfw is double buffered
			renderCount = 2
		}
		if renderCount > 0 {
			coreRenderer.Render(canvas)
			romRenderer.Render(canvas)
			led0Renderer.Render(canvas)
			canvas.SetFillStyle("#ccc")
			canvas.FillRect(20, float64(canvas.Height())-70, float64(canvas.Width()), 80)
			canvas.SetFillStyle("#000")
			canvas.FillText(fmt.Sprintf("FPS=%3.1f, CPU Clock=%3.2f kHz",
				wnd.FPS(), (wnd.FPS()*float32(clocksPerRender))/1000),
				20, float64(canvas.Height())-40)

			canvas.FillText(fmt.Sprintf("'C'=Step Clock 'S'=Step Cycle 'R'=Free Run 'Q'=Quit"),
				20, float64(canvas.Height())-10)
			renderCount--
		}
		currentRunFlags.StepClock = false
		if cycleCount > 7 {
			currentRunFlags.StepCycle = false
			cycleCount = 0
		}
	})

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
