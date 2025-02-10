# Intel 8086 Simulator
[![Go](https://github.com/P100sch/Intel8086Simulator/actions/workflows/go.yml/badge.svg)](https://github.com/P100sch/Intel8086Simulator/actions/workflows/go.yml)

Simple 8086 simulator for Casey Muratori's Performance Programming course.

## Usage

`Intel8086Simulator instructions.bin [-o out.asm]`

Outputs a disassembly of the instruction stream to the console or to the file specified by the `-o` flag.

## Testing

Requires [NASM](https://www.nasm.us) to be installed.
