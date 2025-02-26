package Simulation

import (
	"log"
	"strconv"

	"github.com/P100sch/Intel8086Simulator/Simulation/Shared"
)

//region Types

type DecodingError struct {
	Message string
	Pos     int
}

func (e *DecodingError) Error() string {
	return "Position " + strconv.FormatInt(int64(e.Pos), 16) + ": " + e.Message
}

func newInvalidParameterError(segment, offset uint16, cause string) *DecodingError {
	return &DecodingError{Message: "invalid parameters (" + cause + ")", Pos: convertVirtualAddress(segment, offset)}
}

func newInvalidParameterErrorInvalidInstruction(segment, offset uint16) *DecodingError {
	return newInvalidParameterError(segment, offset, "invalid instruction in register portion")
}

func newUnsupportedError(segment, offset uint16, reason string) *DecodingError {
	return &DecodingError{Message: "unsupported function (" + reason + ")", Pos: convertVirtualAddress(segment, offset)}
}

//endregion

// Simulate reads instruction stream and simulates execution
// Possible errors:
//   - invalid instruction
//   - invalid parameters
//   - instruction stream stops before complete decoding of instruction
//
//goland:noinspection SpellCheckingInspection
func Simulate(logger *log.Logger) error {
	var startOfInstruction = IP
	var baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles int

	for {
		currentInstructionByte := readCodeB(IP)

		switch currentInstructionByte {

		//Standard MOV permutations
		case 0b10001000:
			fallthrough
		case 0b10001001:
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			sourceValue := readRegister(wide | parameter&Shared.RegMask>>3)
			segment, offset := calculateSegmentAndDisplacementByParameter(parameter, IP)
			writeRMValue(parameter, segment, offset, sourceValue, wide)
			IP = incrementIPByParameter(IP, parameter)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, true, wide != 0, true, 2, 8, 9)
		case 0b10001010:
			fallthrough
		case 0b10001011:
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			sourceValue, _, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, wide)
			writeRegister(wide|parameter&Shared.RegMask>>3, sourceValue)
			IP = incrementIPByParameter(IP, parameter)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, false, wide != 0, true, 2, 8, 9)
		//MOV segment register to R/M
		case 0b10001100:
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			var sourceValue uint16
			switch parameter & Shared.RegMask {
			case 0b000000:
				sourceValue = ES
			case 0b001000:
				sourceValue = CS
			case 0b010000:
				sourceValue = SS
			case 0b011000:
				sourceValue = DS
			default:
				return newInvalidParameterErrorInvalidInstruction(CS, IP)
			}
			segment, offset := calculateSegmentAndDisplacementByParameter(parameter, IP)
			writeRMValue(parameter, segment, offset, sourceValue, Shared.WIDE)
			IP = incrementIPByParameter(IP, parameter)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, true, true, true, 2, 0, 8)
		//MOV R/M to segment register
		case 0b10001110:
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			sourceValue, _, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, Shared.WIDE)
			IP = incrementIPByParameter(IP, parameter)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, false, true, true, 2, 9, 0)
			switch parameter & Shared.RegMask {
			case 0b000000:
				ES = sourceValue
			case 0b001000:
				instruction := readInstruction(startOfInstruction, IP)
				CS = sourceValue
				IP = wrapIncrement(IP)
				totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
				logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
				startOfInstruction = IP
				continue
			case 0b010000:
				SS = sourceValue
			case 0b011000:
				DS = sourceValue
			default:
				return newInvalidParameterErrorInvalidInstruction(CS, IP)
			}
		//MOV immediate
		case 0b11000110:
			fallthrough
		case 0b11000111:
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			if parameter&Shared.RegMask != 0 {
				return newInvalidParameterErrorInvalidInstruction(CS, IP)
			}
			segment, offset := calculateSegmentAndDisplacementByParameter(parameter, IP)
			IP = wrapIncrement(incrementIPByParameter(IP, parameter))
			immediate := readCode(IP, wide != 0)
			if wide != 0 {
				IP = wrapIncrement(IP)
			}
			writeRMValue(parameter, segment, offset, immediate, wide)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, true, wide != 0, true, 4, 0, 10)
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
			register := currentInstructionByte & 0b00001111
			IP = wrapIncrement(IP)
			sourceValue := readCode(IP, register > 0b0111)
			if register > 0b0111 {
				IP = wrapIncrement(IP)
			}
			writeRegister(register, sourceValue)
			baseClockCycles, decodingCycles, penaltyCycles = 4, 0, 0

		//ADD/OR/ADC/SUB/AND/SBB/CMP immediate to R/M
		case 0b10000000:
			fallthrough
		case 0b10000001:
			fallthrough
		case 0b10000010:
			fallthrough
		case 0b10000011:
			signExtended := currentInstructionByte&Shared.DirectionMask != 0
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			sourceValue, segment, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, wide)
			IP = wrapIncrement(incrementIPByParameter(IP, parameter))
			immediate := readCode(IP, wide != 0 && !signExtended)
			if wide != 0 && !signExtended {
				IP = wrapIncrement(IP)
			}
			if signExtended {
				immediate = signExtend(immediate)
			}
			var result = sourceValue
			memoryCycles := 17
			//var name = [8]string{"ADD ", "OR ", "ADC ", "SBB ", "AND ", "SUB ", "XOR ", "CMP "}[parameter&Shared.RegMask>>3]
			switch parameter & Shared.RegMask {
			case 0b000000:
				result = addAndUpdateFlags(sourceValue, immediate, wide != 0)
			case 0b101000:
				result = subAndUpateFlags(sourceValue, immediate, wide != 0)
			case 0b111000:
				_ = subAndUpateFlags(sourceValue, immediate, wide != 0)
				memoryCycles = 10
			default:
				return newUnsupportedError(CS, IP, "operation not implemented")
			}
			writeRMValue(parameter, segment, offset, result, wide)
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, true, wide != 0, false, 4, 0, memoryCycles)

		//ADD register with R/M
		case 0b00000000:
			fallthrough
		case 0b00000001:
			fallthrough
		case 0b00000010:
			fallthrough
		case 0b00000011:
			sourceInReg := currentInstructionByte&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			reg := wide | (parameter & Shared.RegMask >> 3)
			regValue := readRegister(reg)
			rmValue, segment, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, wide)
			IP = incrementIPByParameter(IP, parameter)
			sum := addAndUpdateFlags(regValue, rmValue, wide != 0)
			if sourceInReg {
				writeRMValue(parameter, segment, offset, sum, wide)
			} else {
				writeRegister(reg, sum)
			}
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, sourceInReg, wide != 0, false, 3, 9, 16)
		//ADD immediate to accumulator
		case 0b00000100:
			fallthrough
		case 0b00000101:
			wide := currentInstructionByte&Shared.WideMask != 0
			IP = wrapIncrement(IP)
			sourceValue := readCode(IP, wide)
			if wide {
				IP = wrapIncrement(IP)
			}
			sum := addAndUpdateFlags(AX, sourceValue, wide)
			if wide {
				AX = sum
			} else {
				AX = writeL(AX, sum)
			}
			baseClockCycles, decodingCycles, penaltyCycles = 4, 0, 0

		//SUB register and R/M
		case 0b00101000:
			fallthrough
		case 0b00101001:
			fallthrough
		case 0b00101010:
			fallthrough
		case 0b00101011:
			sourceInReg := currentInstructionByte&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			reg := wide | (parameter & Shared.RegMask >> 3)
			regValue := readRegister(reg)
			rmValue, segment, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, wide)
			IP = incrementIPByParameter(IP, parameter)
			if sourceInReg {
				writeRMValue(parameter, segment, offset, subAndUpateFlags(rmValue, regValue, wide != 0), wide)
			} else {
				writeRegister(reg, subAndUpateFlags(regValue, rmValue, wide != 0))
			}
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, sourceInReg, wide != 0, false, 3, 9, 16)
		//SUB immediate from accumulator
		case 0b00101100:
			fallthrough
		case 0b00101101:
			wide := currentInstructionByte&Shared.WideMask != 0
			IP = wrapIncrement(IP)
			sourceValue := readCode(IP, wide)
			if wide {
				IP = wrapIncrement(IP)
			}
			difference := subAndUpateFlags(AX, sourceValue, wide)
			if wide {
				AX = difference
			} else {
				AX = writeL(AX, difference)
			}
			baseClockCycles, decodingCycles, penaltyCycles = 4, 0, 0

		//CMP register to R/M
		case 0b00111000:
			fallthrough
		case 0b00111001:
			fallthrough
		case 0b00111010:
			fallthrough
		case 0b00111011:
			sourceInReg := currentInstructionByte&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(currentInstructionByte)
			IP = wrapIncrement(IP)
			parameter := readCodeB(IP)
			regValue := readRegister(wide | (parameter & Shared.RegMask >> 3))
			rmValue, _, offset := readRMValueSegmentAndDisplacementByParameter(parameter, IP, wide)
			if sourceInReg {
				_ = subAndUpateFlags(rmValue, regValue, wide != 0)
			} else {
				_ = subAndUpateFlags(regValue, rmValue, wide != 0)
			}
			baseClockCycles, decodingCycles, penaltyCycles = getBaseDecodingAndPenaltyCyclesByParameter(parameter, offset, sourceInReg, wide != 0, true, 3, 9, 9)
		//CMP immediate with accumulator
		case 0b00111100:
			fallthrough
		case 0b00111101:
			wide := currentInstructionByte&Shared.WideMask != 0
			IP = wrapIncrement(IP)
			sourceValue := readCode(IP, wide)
			if wide {
				IP = wrapIncrement(IP)
			}
			_ = subAndUpateFlags(AX, sourceValue, wide)
			baseClockCycles, decodingCycles, penaltyCycles = 4, 0, 0

		//JMP
		case 0b11101001:
			IP = wrapIncrement(IP)
			offset := readCodeW(IP)
			IP = wrapAdd(IP, 2)
			instruction := readInstruction(startOfInstruction, IP)
			IP = calculateJump(offset, IP)
			baseClockCycles, decodingCycles, penaltyCycles = 15, 0, 0
			totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
			logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
			startOfInstruction = IP
			continue
		case 0b11101010:
			IP = wrapIncrement(IP)
			newIP := readCodeW(IP)
			IP = wrapAdd(IP, 2)
			newCS := readCodeW(IP)
			IP = wrapAdd(IP, 1)
			instruction := readInstruction(startOfInstruction, IP)
			CS = newCS
			IP = newIP
			baseClockCycles, decodingCycles, penaltyCycles = 15, 0, 0
			totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
			logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
			startOfInstruction = IP
			continue
		//JMP byte
		case 0b11101011:
			IP = wrapIncrement(IP)
			instruction := readInstruction(startOfInstruction, IP)
			IP = calculateJumpB(readCodeB(IP), IP)
			baseClockCycles, decodingCycles, penaltyCycles = 15, 0, 0
			totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
			logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
			startOfInstruction = IP
			continue
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
			condition := conditions[currentInstructionByte&0b00001111]
			IP = wrapIncrement(IP)
			baseClockCycles, decodingCycles, penaltyCycles = 4, 0, 0
			if condition != 0 {
				instruction := readInstruction(startOfInstruction, IP)
				IP = calculateJumpB(readCodeB(IP), IP)
				baseClockCycles = 16
				totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
				logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
				startOfInstruction = IP
				continue
			}
		//LOOP/LOOPZ/LOOPNZ --CX times
		case 0b11100010:
			fallthrough
		case 0b11100001:
			fallthrough
		case 0b11100000:
			condition := [3]byte{ZF ^ 1, ZF, 1}[currentInstructionByte&0b00000011]
			IP = wrapIncrement(IP)
			baseClockCycles, decodingCycles, penaltyCycles = [3]int{5, 6, 5}[currentInstructionByte&0b00000011], 0, 0
			CX--
			if CX > 0 && condition != 0 {
				instruction := readInstruction(startOfInstruction, IP)
				IP = calculateJumpB(readCodeB(IP), IP)
				baseClockCycles = [3]int{19, 18, 17}[currentInstructionByte&0b00000011]
				totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
				logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
				startOfInstruction = IP
				continue
			}
		//JCXZ
		case 0b11100011:
			IP = wrapIncrement(IP)
			baseClockCycles, decodingCycles, penaltyCycles = 6, 0, 0
			if CX == 0 {
				instruction := readInstruction(startOfInstruction, IP)
				IP = calculateJumpB(readCodeB(IP), IP)
				baseClockCycles = 18
				totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
				logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
				startOfInstruction = IP
				continue
			}

		//HLT
		case 0b11110100:
			baseClockCycles, decodingCycles, penaltyCycles = 2, 0, 0
			totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
			logStateAndInstruction(readInstruction(startOfInstruction, IP), baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
			return nil

		default:
			return newUnsupportedError(CS, IP, "unsupported instruction")
		}

		instruction := readInstruction(startOfInstruction, IP)
		IP = wrapIncrement(IP)
		totalClockCycles += baseClockCycles + decodingCycles + penaltyCycles
		logStateAndInstruction(instruction, baseClockCycles, decodingCycles, penaltyCycles, totalClockCycles, logger)
		startOfInstruction = IP
	}
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
	CS = RESET_CS
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

	clear(Memory[:])
}
