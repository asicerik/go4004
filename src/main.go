package main

import (
	"cpucore"
	"css"
	"fmt"
	"image"
	"time"

	"github.com/tfriedel6/canvas/glfwcanvas"
)

func main() {
	fmt.Println("Welcome to the go 4004 emulator :)")

	wnd, canvas, err := glfwcanvas.CreateWindow(1024, 768, "Go 4004")
	if err != nil {
		fmt.Println("Could not open Window")
		return
	}
	canvas.SetFont("C:\\Windows\\Fonts\\courbd.ttf", 24)
	defer wnd.Close()

	core := cpucore.Core{}
	core.Init()
	coreRenderer := cpucore.Renderer{}
	coreRenderer.InitRender(&core, canvas, image.Rectangle{
		image.Point{int(css.Margin), int(css.Margin)},
		image.Point{canvas.Width() - int(2*css.Margin), canvas.Height() - int(2*css.Margin)}})

	renderCount := 0
	lastTime := time.Now()
	wnd.MainLoop(func() {
		currTime := time.Now()
		if currTime.Sub(lastTime).Seconds() >= 0.01 {
			lastTime = currTime
			core.Step()
			// Render twice because glfw is double buffered
			renderCount = 2
		}
		if renderCount > 0 {
			coreRenderer.Render(canvas)
			canvas.SetFillStyle("#ccc")
			canvas.FillRect(20, float64(canvas.Height())-70, 200, 40)
			canvas.SetFillStyle("#000")
			canvas.FillText(fmt.Sprintf("FPS=%3.1f", wnd.FPS()), 20, float64(canvas.Height())-40)
			renderCount--
		}
	})

	fmt.Println("Goodbye")
}
