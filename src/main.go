package main

import (
	"cpucore"
	"fmt"
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
	canvas.SetFont("C:\\Windows\\Fonts\\cour.ttf", 24)
	defer wnd.Close()

	core := cpucore.Core{}
	core.Init()
	// core.Test()

	lastTime := time.Now()
	wnd.MainLoop(func() {
		currTime := time.Now()
		if currTime.Sub(lastTime).Seconds() >= 1.0 {
			lastTime = currTime
			core.Step()
		}
		core.Render(canvas)
	})

	fmt.Println("Goodbye")
}
