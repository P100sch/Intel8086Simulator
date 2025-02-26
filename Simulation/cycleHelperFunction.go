package Simulation

import "github.com/P100sch/Intel8086Simulator/Simulation/Shared"

func getBaseDecodingAndPenaltyCyclesByParameter(parameter byte, virtualDataAddress uint16, sourceInReg, wide, readOnce bool, regCycles, fromMemoryCycles, toMemoryCycles int) (baseCycles, decodingCycles, penaltyCycles int) {
	if parameter&Shared.ModMask == Shared.RegisterMode {
		baseCycles = regCycles
		return
	}
	if wide {
		penaltyCycles = int(virtualDataAddress&1) * 4
	}
	if sourceInReg {
		baseCycles = toMemoryCycles
		if !readOnce {
			penaltyCycles *= 2
		}
	} else {
		baseCycles = fromMemoryCycles
	}
	mod := parameter & Shared.ModMask
	switch parameter & Shared.RMMask {
	case 0b000000:
		fallthrough
	case 0b011:
		decodingCycles = 7
	case 0b001:
		fallthrough
	case 0b010:
		decodingCycles = 8
	case 0b100:
		fallthrough
	case 0b101:
		fallthrough
	case 0b111:
		decodingCycles = 5
	case 0b110:
		if mod == Shared.MemoryMode {
			decodingCycles = 6
		} else {
			decodingCycles = 5
		}
	}
	if mod != Shared.MemoryMode {
		decodingCycles += 4
	}
	return
}
