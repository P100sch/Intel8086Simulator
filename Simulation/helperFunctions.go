package Simulation

const (
	_L      uint16 = 0b0000000011111111
	_H      uint16 = 0b1111111100000000
	_W_SIGN uint16 = 0b1000000000000000
	_B_SIGN uint16 = 0b0000000010000000
	_W_MAX  uint16 = 0b1111111111111111
	_B_MAX  byte   = 0b11111111
)

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

func readWord(data []byte, position int) uint16 {
	return uint16(data[position-1]) | uint16(data[position])<<8
}

func read(data []byte, position *int, wide bool) (uint16, error) {
	*position++
	if wide {
		*position++
	}
	if *position >= len(data) {
		return 0, newInvalidParameterErrorPrematureEndOfStream(*position)
	}
	if wide {
		return readWord(data, *position), nil
	}
	return uint16(data[*position]), nil
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

func calculateJump(offset uint16, position int) uint16 {
	return uint16((uint32(position+1) + uint32(offset)) & uint32(_W_MAX))
}

func calculateJumpB(offset uint8, position int) uint16 {
	return calculateJump(signExtend(uint16(offset)), position)
}
