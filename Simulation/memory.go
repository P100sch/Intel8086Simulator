package Simulation

var AX uint16
var BX uint16
var CX uint16
var DX uint16
var SP uint16
var BP uint16
var SI uint16
var DI uint16
var IP uint16

var CS = RESET_CS
var DS uint16
var SS uint16
var ES uint16

var TF byte
var DF byte
var IF byte
var OF byte
var SF byte
var ZF byte
var AF byte
var PF byte
var CF byte

var Memory = [0xFFFFFF]byte{}

func convertVirtualAddress(segment uint16, offset uint16) int {
	return (int(segment)<<4 + int(offset)) & 0xFFFFFF
}

func read(segment, offset uint16, wide bool) uint16 {
	if wide {
		return readW(segment, offset)
	}
	return uint16(Memory[convertVirtualAddress(segment, offset)])
}

func readW(segment, offset uint16) uint16 {
	return uint16(Memory[convertVirtualAddress(segment, offset)]) | uint16(Memory[convertVirtualAddress(segment, wrapIncrement(offset))])<<8
}

func readCode(offset uint16, wide bool) uint16 {
	if wide {
		return readCodeW(offset)
	}
	return uint16(readCodeB(offset))
}

func readCodeB(offset uint16) byte {
	return Memory[convertVirtualAddress(CS, offset)]
}

func readCodeW(offset uint16) uint16 {
	return uint16(readCodeB(offset)) | uint16(readCodeB(wrapIncrement(offset)))<<8
}

func readData(offset uint16, wide bool) uint16 {
	if wide {
		return readDataW(offset)
	}
	return uint16(readDataB(offset))
}

func readDataB(offset uint16) byte {
	return Memory[convertVirtualAddress(DS, offset)]
}

func readDataW(offset uint16) uint16 {
	return uint16(readDataB(offset)) | uint16(readDataB(wrapIncrement(offset)))<<8
}

func write(segment, offset, value uint16, wide bool) {
	Memory[convertVirtualAddress(segment, offset)] = byte(value & uint16(_B_MAX))
	if wide {
		Memory[convertVirtualAddress(segment, wrapIncrement(offset))] = byte(value >> 8)
	}
}

type MemoryWriteError string

func (e MemoryWriteError) Error() string {
	return "memory write error: " + string(e)
}

func LoadProgram(data []byte, isIncomplete bool) error {
	if len(data) > len(Memory) {
		return MemoryWriteError("program too big")
	}
	copy(Memory[:], data)
	if isIncomplete {
		const RESET_VECTOR = int(RESET_CS) << 4
		if len(data) < RESET_VECTOR-1 && data[len(data)-1] != 0b11110100 {
			//HLT
			Memory[len(data)] = 0b11110100
		}
		if len(data) < RESET_VECTOR {
			//JMP
			Memory[RESET_VECTOR] = 0b11101010
			Memory[RESET_VECTOR+1] = 0
			Memory[RESET_VECTOR+2] = 0
			Memory[RESET_VECTOR+3] = 0
			Memory[RESET_VECTOR+4] = 0
		}
	}
	return nil
}
