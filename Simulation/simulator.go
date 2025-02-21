package Simulation

import (
	"log"
	"strconv"

	"github.com/P100sch/Intel8086Simulator/Simulation/Disassembly"
	"github.com/P100sch/Intel8086Simulator/Simulation/Shared"
)

//region Types

type DecodingError struct {
	Message string
	Pos     int
}

func (e *DecodingError) Error() string {
	return "Position " + strconv.Itoa(e.Pos) + ": " + e.Message
}

func newInvalidParameterError(position int, cause string) *DecodingError {
	return &DecodingError{Message: "invalid parameters (" + cause + ")", Pos: position}
}

func newInvalidParameterErrorPrematureEndOfStream(position int) *DecodingError {
	return newInvalidParameterError(position, "reached end of instruction stream while decoding")
}

func newInvalidParameterErrorInvalidInstruction(position int) *DecodingError {
	return newInvalidParameterError(position, "invalid instruction in register portion")
}

func newUnsupportedError(position int, reason string) *DecodingError {
	return &DecodingError{Message: "unsupported function (" + reason + ")", Pos: position}
}

//endregion

// Simulate reads instruction stream and simulates execution
// Possible errors:
//   - invalid instruction
//   - invalid parameters
//   - instruction stream stops before complete decoding of instruction
//
//goland:noinspection SpellCheckingInspection
func Simulate(fileData []byte, logger *log.Logger) error {
	var err error
	dataLength := len(fileData)
	startOfInstruction := 0

	var sourceValue uint16

	for position := 0; position < dataLength; position++ {
		var ipModified = false

		switch fileData[position] {

		//Standard MOV permutations
		case 0b10001000:
			fallthrough
		case 0b10001001:
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			sourceValue = readRegister(wide | fileData[position]&Shared.RegMask>>3)
			if fileData[position]&Shared.ModMask == Shared.RegisterMode {
				writeRegister(wide|fileData[position]&Shared.RMMask, sourceValue)
			} else {
				return newUnsupportedError(position, "memory not implemented")
			}
		case 0b10001010:
			fallthrough
		case 0b10001011:
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.ModMask == Shared.RegisterMode {
				sourceValue = readRegister(wide | fileData[position]&Shared.RMMask)
			} else {
				return newUnsupportedError(position, "memory not implemented")
			}
			writeRegister(wide|fileData[position]&Shared.RegMask>>3, sourceValue)
		//MOV segment register to R/M
		case 0b10001100:
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			switch fileData[position] & Shared.RegMask {
			case 0b000000:
				sourceValue = ES
			case 0b001000:
				sourceValue = CS
			case 0b010000:
				sourceValue = SS
			case 0b011000:
				sourceValue = DS
			default:
				return newInvalidParameterErrorInvalidInstruction(position)
			}
			if fileData[position]&Shared.ModMask == Shared.RegisterMode {
				writeRegister(Shared.WIDE|fileData[position]&Shared.RMMask, sourceValue)
			} else {
				return newUnsupportedError(position, "memory not implemented")
			}
		//MOV R/M to segment register
		case 0b10001110:
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.ModMask == Shared.RegisterMode {
				sourceValue = readRegister(Shared.WIDE | fileData[position]&Shared.RMMask)
			} else {
				return newUnsupportedError(position, "memory not implemented")
			}
			switch fileData[position] & Shared.RegMask {
			case 0b000000:
				ES = sourceValue
			case 0b001000:
				CS = sourceValue
			case 0b010000:
				SS = sourceValue
			case 0b011000:
				DS = sourceValue
			default:
				return newInvalidParameterErrorInvalidInstruction(position)
			}
		//MOV immediate
		case 0b11000110:
			fallthrough
		case 0b11000111:
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.RegMask != 0 {
				return newInvalidParameterErrorInvalidInstruction(position)
			}
			rmBits := fileData[position] & Shared.RMMask
			sourceValue, err = read(fileData, &position, wide != 0)
			if err != nil {
				return err
			}
			if fileData[position]&Shared.ModMask == Shared.RegisterMode {
				writeRegister(wide|rmBits, sourceValue)
			} else {
				return newUnsupportedError(position, "memory not implemented")
			}
		//MOV immediate into register
		case 0b10110000:
			fallthrough
		case 0b10110001:
			fallthrough
		case 0b10110010:
			fallthrough
		case 0b10110011:
			fallthrough
		case 0b10110100:
			fallthrough
		case 0b10110101:
			fallthrough
		case 0b10110110:
			fallthrough
		case 0b10110111:
			fallthrough
		case 0b10111000:
			fallthrough
		case 0b10111001:
			fallthrough
		case 0b10111010:
			fallthrough
		case 0b10111011:
			fallthrough
		case 0b10111100:
			fallthrough
		case 0b10111101:
			fallthrough
		case 0b10111110:
			fallthrough
		case 0b10111111:
			register := fileData[position] & 0b00001111
			sourceValue, err = read(fileData, &position, register > 0b0111)
			if err != nil {
				return err
			}
			writeRegister(register, sourceValue)

		//ADD/OR/ADC/SUB/AND/SBB/CMP immediate to R/M
		case 0b10000000:
			fallthrough
		case 0b10000001:
			fallthrough
		case 0b10000010:
			fallthrough
		case 0b10000011:
			signExtended := fileData[position]&Shared.DirectionMask != 0
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			operation := fileData[position] & Shared.RegMask
			register := fileData[position] & Shared.RMMask
			if fileData[position]&Shared.ModMask != Shared.RegisterMode {
				return newUnsupportedError(position, "memory not implemented")
			}
			var immediate uint16
			immediate, err = read(fileData, &position, wide != 0 && !signExtended)
			if err != nil {
				return err
			}
			if signExtended {
				immediate = signExtend(immediate)
			}
			//var name = [8]string{"ADD ", "OR ", "ADC ", "SBB ", "AND ", "SUB ", "XOR ", "CMP "}[fileData[position]&Shared.RegMask>>3]
			switch operation {
			case 0b000000:
				writeRegister(wide|register, add(readRegister(wide|register), immediate, wide != 0))
			case 0b101000:
				writeRegister(wide|register, sub(readRegister(wide|register), immediate, wide != 0))
			case 0b111000:
				_ = sub(readRegister(wide|register), immediate, wide != 0)
			default:
				return newUnsupportedError(position, "operation not implemented")
			}

		//ADD register with R/M
		case 0b00000000:
			fallthrough
		case 0b00000001:
			fallthrough
		case 0b00000010:
			fallthrough
		case 0b00000011:
			sourceInReg := fileData[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.ModMask != Shared.RegisterMode {
				return newUnsupportedError(position, "memory not implemented")
			}
			reg := wide | (fileData[position] & Shared.RegMask >> 3)
			regValue := readRegister(reg)
			rmRegister := wide | fileData[position]&Shared.RMMask
			rmValue := readRegister(wide | rmRegister)
			sum := add(regValue, rmValue, wide != 0)
			if sourceInReg {
				writeRegister(rmRegister, sum)
			} else {
				writeRegister(reg, sum)
			}
		//ADD immediate to accumulator
		case 0b00000100:
			fallthrough
		case 0b00000101:
			wide := fileData[position]&Shared.WideMask != 0
			sourceValue, err = read(fileData, &position, wide)
			if err != nil {
				return err
			}
			sum := add(AX, sourceValue, wide)
			if wide {
				AX = sum
			} else {
				AX = writeL(AX, sum)
			}

		//SUB register and R/M
		case 0b00101000:
			fallthrough
		case 0b00101001:
			fallthrough
		case 0b00101010:
			fallthrough
		case 0b00101011:
			sourceInReg := fileData[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.ModMask != Shared.RegisterMode {
				return newUnsupportedError(position, "memory not implemented")
			}
			reg := wide | (fileData[position] & Shared.RegMask >> 3)
			regValue := readRegister(reg)
			rmRegister := wide | fileData[position]&Shared.RMMask
			rmValue := readRegister(wide | rmRegister)
			if sourceInReg {
				writeRegister(rmRegister, sub(rmValue, regValue, wide != 0))
			} else {
				writeRegister(reg, sub(regValue, rmValue, wide != 0))
			}
		//SUB immediate from accumulator
		case 0b00101100:
			fallthrough
		case 0b00101101:
			wide := fileData[position]&Shared.WideMask != 0
			sourceValue, err = read(fileData, &position, wide)
			if err != nil {
				return err
			}
			difference := sub(AX, sourceValue, wide)
			if wide {
				AX = difference
			} else {
				AX = writeL(AX, difference)
			}

		//CMP register to R/M
		case 0b00111000:
			fallthrough
		case 0b00111001:
			fallthrough
		case 0b00111010:
			fallthrough
		case 0b00111011:
			sourceInReg := fileData[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(fileData[position])
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if fileData[position]&Shared.ModMask != Shared.RegisterMode {
				return newUnsupportedError(position, "memory not implemented")
			}
			regValue := readRegister(wide | (fileData[position] & Shared.RegMask >> 3))
			rmRegister := wide | fileData[position]&Shared.RMMask
			rmValue := readRegister(wide | rmRegister)
			sourceValue = readRegister(wide | rmRegister)
			if sourceInReg {
				_ = sub(rmValue, regValue, wide != 0)
			} else {
				_ = sub(regValue, rmValue, wide != 0)
			}
		//CMP immediate with accumulator
		case 0b00111100:
			fallthrough
		case 0b00111101:
			wide := fileData[position]&Shared.WideMask != 0
			sourceValue, err = read(fileData, &position, wide)
			if err != nil {
				return err
			}
			_ = sub(AX, sourceValue, wide)

		//JMP
		case 0b11101001:
			position += 2
			if position >= dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			IP = calculateJump(uint16(fileData[position-1])+uint16(fileData[position])<<8, position)
			ipModified = true
		case 0b11101010:
			position += 4
			if position >= dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			IP = uint16(fileData[position-3]) + uint16(fileData[position-2])<<8
			CS = uint16(fileData[position-1]) + uint16(fileData[position-2])<<8
			ipModified = true
		//JMP byte
		case 0b11101011:
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			IP = calculateJumpB(fileData[position], position)
			ipModified = true
		//Conditional jumps
		case 0b01110100:
			fallthrough
		case 0b01111100:
			fallthrough
		case 0b01111110:
			fallthrough
		case 0b01110010:
			fallthrough
		case 0b01110110:
			fallthrough
		case 0b01111010:
			fallthrough
		case 0b01110000:
			fallthrough
		case 0b01111000:
			fallthrough
		case 0b01110101:
			fallthrough
		case 0b01111101:
			fallthrough
		case 0b01111111:
			fallthrough
		case 0b01110011:
			fallthrough
		case 0b01110111:
			fallthrough
		case 0b01111011:
			fallthrough
		case 0b01110001:
			fallthrough
		case 0b01111001:
			var conditions = [...]uint16{
				0b0100: uint16(ZF),                   //JZ
				0b1100: uint16(SF ^ OF),              //JL
				0b1110: uint16(SF ^ OF | ZF&^OF),     //JLE
				0b0010: uint16(CF),                   //JB
				0b0110: uint16(CF | ZF),              //JBE
				0b1010: uint16(PF),                   //JP
				0b0000: uint16(OF),                   //JO
				0b1000: uint16(SF),                   //JS
				0b0101: uint16(ZF ^ 1),               //JNZ
				0b1101: uint16(SF ^ OF ^ 1 | ZF&^OF), //JGE
				0b1111: uint16(SF ^ OF ^ 1),          //JG
				0b0011: uint16(CF ^ 1 | ZF),          //JAE
				0b0111: uint16(CF ^ 1),               //JA
				0b1011: uint16(PF ^ 1),               //JNP
				0b0001: uint16(OF ^ 1),               //JNO
				0b1001: uint16(SF ^ 1),               //JNS
			}
			condition := conditions[fileData[position]&0b00001111]
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if condition != 0 {
				IP = calculateJumpB(fileData[position], position)
				ipModified = true
			}
		//LOOP/LOOPZ/LOOPNZ --CX times
		case 0b11100010:
			fallthrough
		case 0b11100001:
			fallthrough
		case 0b11100000:
			condition := [3]byte{ZF ^ 1, ZF, 1}[fileData[position]&0b00000011]
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			CX--
			if CX > 0 && condition != 0 {
				IP = calculateJumpB(fileData[position], position)
				ipModified = true
			}
		//JXCZ
		case 0b11100011:
			position++
			if position == dataLength {
				return newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if CX == 0 {
				IP = calculateJumpB(fileData[position], position)
				ipModified = true
			}

		default:
			return newUnsupportedError(position, "unsupported instruction")
		}

		if ipModified {
			logStateAndInstruction(fileData[startOfInstruction:position+1], logger)
			position = int(IP - 1)
		} else {
			IP = uint16(position + 1)
			logStateAndInstruction(fileData[startOfInstruction:position+1], logger)
		}
		startOfInstruction = position + 1
	}

	return nil
}

func Rest() {
	AX = 0
	BX = 0
	CX = 0
	DX = 0
	SP = 0
	BP = 0
	SI = 0
	DI = 0
	IP = 0
	CS = 0
	DS = 0
	SS = 0
	ES = 0
	TF = 0
	DF = 0
	IF = 0
	OF = 0
	SF = 0
	ZF = 0
	AF = 0
	PF = 0
	CF = 0
}

func logStateAndInstruction(instruction []byte, logger *log.Logger) {
	if logger != nil {
		assembly, err := Disassembly.Disassemble(instruction)
		if err != nil {
			logger.Println(err.Error())
		} else {
			logger.Println(formatState() + " ; " + assembly)
		}
	}
}
