# go4004
An Intel 4004 CPU emulator
This was a fun little project blending two of my skills, hardware and software. I used to be an ASIC engineer (could you have guessed from my username)? I thought it would be fun to write a near cycle-accurate model of the venerable Intel 4004 microprocessor; the first commercially available microprocessor from Intel.
This project is written in Go, with an optional graphics front-end using SDL. The graphics allowed me to visualize the processor in action. I modeled the UI to look like the block diagram you can find on the internet.

![Visualizer Program in Action](https://dl.dropboxusercontent.com/s/cho3c26wtrkhzh4/Go%204004.jpg?dl=0)

As you can see in this image, the CPU clock is running at about 150kHz. The real CPU was rated at 750kHz, so we are a little off :)
**However**,  the intent of this project was not to make a fast emulator, but rather something that actually models the CPU and its peripherals. 
## Current project status as of 12/31/2018

 - 4004 CPU without the ALU yet. So, all non ALU base instructions are implemented
 - 4001 ROM with I/O read and write ports
 - CPU/ROM/IO visualizer (see picture above)
 - Unit tests for all implemented instructions and ROM
