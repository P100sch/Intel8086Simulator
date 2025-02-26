# Intel 8086 Simulator
[![Go](https://github.com/P100sch/Intel8086Simulator/actions/workflows/go.yml/badge.svg)](https://github.com/P100sch/Intel8086Simulator/actions/workflows/go.yml)

Simple 8086 simulator for Casey Muratori's Performance Programming course.

## Usage

`Intel8086Simulator [-v|d] instructions.bin [-o out.asm/data]`
Simulates the execution of the instruction stream.
 - `-v` Outputs disassembly and the state of the registers after each instruction.
 - `-d` Only outputs a disassembly of the instruction stream to the console or to the file specified by the `-o` flag.
 - `-o` saves the final state of memory to the specified file.

## Testing

Requires [NASM](https://www.nasm.us) to be installed.
