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

		default:
			return newUnsupportedError(position, "unsupported instruction")
			//return &DecodingError{Message: "invalid instruction", Pos: position}
		}

		IP = uint16(position + 1)
		logStateAndInstruction(fileData[startOfInstruction:position+1], logger)
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
