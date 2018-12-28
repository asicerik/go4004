package main

import (
	"cpucore"
	"fmt"
)

func main() {
	fmt.Println("Welcome to the go 4004 emulator :)")

	core := cpucore.Core{}
	core.Init()
	core.Test()

	fmt.Println("Goodbye")
}
