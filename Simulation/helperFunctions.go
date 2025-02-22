package Simulation

import "github.com/P100sch/Intel8086Simulator/Simulation/Shared"

const (
	_L      uint16 = 0b0000000011111111
	_H      uint16 = 0b1111111100000000
	_W_SIGN uint16 = 0b1000000000000000
	_B_SIGN uint16 = 0b0000000010000000
	_W_MAX  uint16 = 0b1111111111111111
	_B_MAX  byte   = 0b11111111
)

const RESET_CS uint16 = 0xFFFF

func readH(w uint16) uint16 {
	return w & _H >> 8
}

func writeL(w, b uint16) uint16 {
	return w&_H | b&_L
}

func writeH(w, b uint16) uint16 {
	return w&_L | b<<8
}

func readRegister(register byte) uint16 {
	switch register {
	case 0b0000:
		return AX & _L
	case 0b0001:
		return CX & _L
	case 0b0010:
		return DX & _L
	case 0b0011:
		return BX & _L
	case 0b0100:
		return readH(AX)
	case 0b0101:
		return readH(CX)
	case 0b0110:
		return readH(DX)
	case 0b0111:
		return readH(BX)
	case 0b1000:
		return AX
	case 0b1001:
		return CX
	case 0b1010:
		return DX
	case 0b1011:
		return BX
	case 0b1100:
		return SP
	case 0b1101:
		return BP
	case 0b1110:
		return SI
	case 0b1111:
		return DI
	default:
		panic("Invalid register value")
	}
}

func writeRegister(register byte, value uint16) {
	switch register {
	case 0b0000:
		AX = writeL(AX, value)
	case 0b0001:
		CX = writeL(CX, value)
	case 0b0010:
		DX = writeL(DX, value)
	case 0b0011:
		BX = writeL(BX, value)
	case 0b0100:
		AX = writeH(AX, value)
	case 0b0101:
		CX = writeH(CX, value)
	case 0b0110:
		DX = writeH(DX, value)
	case 0b0111:
		BX = writeH(BX, value)
	case 0b1000:
		AX = value
	case 0b1001:
		CX = value
	case 0b1010:
		DX = value
	case 0b1011:
		BX = value
	case 0b1100:
		SP = value
	case 0b1101:
		BP = value
	case 0b1110:
		SI = value
	case 0b1111:
		DI = value
	default:
		panic("Invalid register value")
	}
}

func signExtend(x uint16) uint16 {
	signExtension := x & 0b10000000 >> 7 * _H
	return signExtension | x
}

func setCommonFlags(value uint16, signBit uint16) {
	if value == 0 {
		ZF = 1
	} else {
		ZF = 0
	}
	odd := (value & 0b11110000 >> 4) ^ (value & 0b1111)
	odd = (odd & 0b1100 >> 2) ^ (odd & 0b11)
	odd = (odd & 0b10 >> 1) ^ (odd & 0b1)
	if odd == 1 {
		PF = 0
	} else {
		PF = 1
	}
	if value&signBit == 0 {
		SF = 0
	} else {
		SF = 1
	}
}

func calculateJump(offset uint16, currentOffset uint16) uint16 {
	return uint16((uint32(currentOffset+1) + uint32(offset)) & uint32(_W_MAX))
}

func calculateJumpB(offset uint8, currentOffset uint16) uint16 {
	return calculateJump(signExtend(uint16(offset)), currentOffset)
}

func calculateSegmentAndDisplacementByParameter(parameter byte, parameterOffset uint16) (segment, displacement uint16) {
	displacementOffset := wrapIncrement(parameterOffset)
	segment = DS
	rm := parameter & Shared.RMMask
	switch rm {
	case 0b000:
		displacement = wrapAdd(BX, SI)
	case 0b001:
		displacement = wrapAdd(BX, DI)
	case 0b010:
		displacement = wrapAdd(BP, SI)
	case 0b011:
		displacement = wrapAdd(BP, DI)
	case 0b100:
		displacement = SI
	case 0b101:
		displacement = DI
	case 0b110:
		displacement = BP
		segment = SS
	case 0b111:
		displacement = BX
	}
	switch parameter & Shared.ModMask {
	case Shared.MemoryMode:
		if rm != 0b110 {
			return
		}
		segment = DS
		displacement = readCodeW(displacementOffset)
		return
	case Shared.Memory8Mode:
		displacement = wrapAdd(displacement, uint16(readCodeB(displacementOffset)))
		return
	case Shared.Memory16Mode:
		displacement = wrapAdd(displacement, readCodeW(displacementOffset))
		return
	case Shared.RegisterMode:
		segment = 0
		displacement = 0
		return
	default:
		panic("impossible state")
	}
}

func incrementIPByParameter(currentIP uint16, parameter byte) uint16 {
	var newIP uint16
	switch parameter & Shared.ModMask {
	case Shared.RegisterMode:
		return currentIP
	case Shared.MemoryMode:
		if parameter&Shared.RMMask != 0b110 {
			return currentIP
		}
		return wrapAdd(currentIP, 2)
	case Shared.Memory8Mode:
		return wrapIncrement(currentIP)
	case Shared.Memory16Mode:
		return wrapAdd(currentIP, 2)
	}
	return newIP
}

func readRMValueSegmentAndDisplacementByParameter(parameter byte, parameterOffset uint16, wide byte) (value, segment, displacement uint16) {
	if parameter&Shared.ModMask != Shared.RegisterMode {
		segment, displacement = calculateSegmentAndDisplacementByParameter(parameter, parameterOffset)
		value = read(segment, displacement, wide != 0)
	} else {
		value = readRegister(wide | parameter&Shared.RMMask)
	}
	return
}

func writeRMValue(parameter byte, segment, displacement, value uint16, wide byte) {
	if parameter&Shared.ModMask != Shared.RegisterMode {
		write(segment, displacement, value, wide != 0)
	} else {
		writeRegister(wide|parameter&Shared.RMMask, value)
	}
}
