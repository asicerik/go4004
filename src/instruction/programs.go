package instruction

func LEDCount() []uint8 {
	data := make([]uint8, 0)
	for i := 0; i < 16; i++ {
		addInstruction(&data, LDM|0)         // Load 5 into the accumulator (chip ID)
		addInstruction(&data, XCH|2)         // Swap accumulator with r2
		addInstruction(&data, LDM|uint8(i))  // Load i value into the accumulator
		addInstruction(&data, FIM_SRC|(2)|1) // Send address in r2,r3 to ROM/RAM
		addInstruction(&data, WRR)           // Write accumulator to ROM
	}
	addInstruction(&data, JUN)  // Jump back to ROM 0
	addInstruction(&data, 0x00) // Jump to address 0
	// Fill the rest of the space up till 256
	zeroes := make([]uint8, 256-len(data))
	data = append(data, zeroes...)
	return data
}

func addInstruction(data *[]uint8, inst uint8) {
	*data = append(*data, inst)
}
