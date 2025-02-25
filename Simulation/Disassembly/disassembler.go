package Disassembly

import (
	"strconv"
	"strings"

	"github.com/P100sch/Intel8086Simulator/Simulation/Shared"
)

//region Types

type DisassembleError struct {
	Message string
	Pos     int
}

func (e *DisassembleError) Error() string {
	return "Position " + strconv.Itoa(e.Pos) + ": " + e.Message
}

func newInvalidParameterError(position int, cause string) *DisassembleError {
	return &DisassembleError{Message: "invalid parameters (" + cause + ")", Pos: position}
}

func newInvalidParameterErrorPrematureEndOfStream(position int) *DisassembleError {
	return newInvalidParameterError(position, "reached end of instruction stream while disassembling")
}

func newInvalidParameterErrorInvalidInstruction(position int) *DisassembleError {
	return newInvalidParameterError(position, "invalid instruction in register portion")
}

//endregion

// Disassemble instruction stream to assembly
// Possible errors:
//   - invalid instruction
//   - invalid parameters
//   - instruction stream stops before complete decoding of instruction
//
//goland:noinspection SpellCheckingInspection
func Disassemble(data []byte) (string, error) {
	startOfInsctruction := -1
	builder := strings.Builder{}

	var err error
	dataLength := len(data)
	var segmentOverride = ""

	for position := 0; position < dataLength; position++ {

		var assembly = ""
		var segmentRegister = false

		switch data[position] {

		//MOV R/M to segment register
		case 0b10001100:
			fallthrough
		//MOV segment register to R/M
		case 0b10001110:
			segmentRegister = true
			fallthrough
		//MOV immediate
		case 0b11000110:
			fallthrough
		case 0b11000111:
			fallthrough
		//Standard MOV permutations
		case 0b10001000:
			fallthrough
		case 0b10001001:
			fallthrough
		case 0b10001010:
			fallthrough
		case 0b10001011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			immediate := data[position]&0b01000000 != 0
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("MOV ", segmentOverride, false, sourceInReg, segmentRegister, immediate, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//MOV accumulator to memory
		case 0b10100000:
			fallthrough
		case 0b10100001:
			fallthrough
		case 0b10100010:
			fallthrough
		case 0b10100011:
			var first string
			if data[position]&0b00000001 == 0 {
				first = "AL"
			} else {
				first = "AX"
			}
			accumulatorIsSource := data[position]&Shared.DirectionMask != 0
			position += 2
			if position >= dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			second := "[" + strconv.FormatUint(uint64(data[position-1])|uint64(data[position])<<8, 10) + "]"
			builder.WriteString("MOV ")
			builder.WriteString(order(accumulatorIsSource, first, second))
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
			builder.WriteString("MOV ")
			builder.WriteString(registers[data[position]&0b00001111])
			builder.WriteString(", ")
			wide := data[position]&0b00001000 != 0
			var value string
			value, err = readData(false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(value)

		//PUSH segment register
		case 0b00000110:
			fallthrough
		case 0b00001110:
			fallthrough
		case 0b00010110:
			fallthrough
		case 0b00011110:
			segmentRegister = true
			fallthrough
		//PUSH register
		case 0b01010000:
			fallthrough
		case 0b01010001:
			fallthrough
		case 0b01010010:
			fallthrough
		case 0b01010011:
			fallthrough
		case 0b01010100:
			fallthrough
		case 0b01010101:
			fallthrough
		case 0b01010110:
			fallthrough
		case 0b01010111:
			builder.WriteString("PUSH ")
			if segmentRegister {
				builder.WriteString(segmentRegisters[data[position]&0b00011000>>3])
			} else {
				builder.WriteString(registers[data[position]&0b00000111|0b00001000])
			}

		//POP R/M
		case 0b10001111:
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if data[position]&Shared.RegMask != 0b00000000 {
				return "", newInvalidParameterErrorInvalidInstruction(position)
			}
			assembly, err = disassembleStandardParameters("POP ", segmentOverride, true, false, false, false, false, Shared.WIDE, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//POP segment register
		case 0b00000111:
			fallthrough
		case 0b00001111:
			fallthrough
		case 0b00010111:
			fallthrough
		case 0b00011111:
			segmentRegister = true
			fallthrough
		//POP register
		case 0b01011000:
			fallthrough
		case 0b01011001:
			fallthrough
		case 0b01011010:
			fallthrough
		case 0b01011011:
			fallthrough
		case 0b01011100:
			fallthrough
		case 0b01011101:
			fallthrough
		case 0b01011110:
			fallthrough
		case 0b01011111:
			builder.WriteString("POP ")
			if segmentRegister {
				builder.WriteString(segmentRegisters[data[position]&0b00011000>>3])
			} else {
				builder.WriteString(registers[data[position]&0b00000111|0b00001000])
			}

		//XCHG register and R/M
		case 0b10000110:
			fallthrough
		case 0b10000111:
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("XCHG ", segmentOverride, false, false, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//XCHG register and accumulator
		case 0b10010000:
			fallthrough
		case 0b10010001:
			fallthrough
		case 0b10010010:
			fallthrough
		case 0b10010011:
			fallthrough
		case 0b10010100:
			fallthrough
		case 0b10010101:
			fallthrough
		case 0b10010110:
			fallthrough
		case 0b10010111:
			builder.WriteString("XCHG AX, ")
			builder.WriteString(registers[data[position]&0b00000111|0b00001000])

		//IN fixed port
		case 0b11100101:
			fallthrough
		case 0b11100100:
			builder.WriteString("IN ")
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("AL, ")
			} else {
				builder.WriteString("AX, ")
			}
			var value string
			value, err = readData(false, false, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(value)
		//OUT fixed port
		case 0b11100111:
			fallthrough
		case 0b11100110:
			wide := data[position]&Shared.WideMask == 0
			builder.WriteString("OUT ")
			var value string
			value, err = readData(false, false, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(value)
			if wide {
				builder.WriteString(", AL\n")
			} else {
				builder.WriteString(", AX\n")
			}
		//IN variable port
		case 0b11101100:
			fallthrough
		case 0b11101101:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("IN AL, DX\n")
			} else {
				builder.WriteString("IN AX, DX\n")
			}
		//OUT variable port
		case 0b11101110:
			fallthrough
		case 0b11101111:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("OUT DX, AL\n")
			} else {
				builder.WriteString("OUT DX, AX\n")
			}

		//LEA load EA to register
		case 0b10001101:
			fallthrough
		//LES load pointer to ES
		case 0b11000100:
			fallthrough
		//LDS load pointer to DS
		case 0b11000101:
			name := [16]string{0b0100: "LES ", 0b0101: "LDS ", 0b1101: "LEA "}[data[position]&0b00001111]
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters(name, segmentOverride, false, false, false, false, false, Shared.WIDE, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//ADD/OR/ADC/SUB/AND/SBB/CMP immediate to R/M
		case 0b10000000:
			fallthrough
		case 0b10000001:
			fallthrough
		case 0b10000010:
			fallthrough
		case 0b10000011:
			signExtend := data[position]&Shared.DirectionMask != 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			var name = [8]string{"ADD ", "OR ", "ADC ", "SBB ", "AND ", "SUB ", "XOR ", "CMP "}[data[position]&Shared.RegMask>>3]
			assembly, err = disassembleStandardParameters(name, segmentOverride, false, false, false, true, signExtend, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//ADD/ADC register with R/M
		case 0b00000000:
			fallthrough
		case 0b00000001:
			fallthrough
		case 0b00000010:
			fallthrough
		case 0b00000011:
			fallthrough
		case 0b00010000:
			fallthrough
		case 0b00010001:
			fallthrough
		case 0b00010010:
			fallthrough
		case 0b00010011:
			var name string
			if data[position]&0b00010000 == 0 {
				name = "ADD "
			} else {
				name = "ADC "
			}
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters(name, segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//ADD immediate to accumulator
		case 0b00000100:
			fallthrough
		case 0b00000101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("ADD ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//ADC immediate to accumulator
		case 0b00010100:
			fallthrough
		case 0b00010101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("ADC ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//SUB/SBB register and R/M
		case 0b00011000:
			fallthrough
		case 0b00011001:
			fallthrough
		case 0b00011010:
			fallthrough
		case 0b00011011:
			fallthrough
		case 0b00101000:
			fallthrough
		case 0b00101001:
			fallthrough
		case 0b00101010:
			fallthrough
		case 0b00101011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			var name string
			if data[position]&0b00010000 == 0 {
				name = "SUB "
			} else {
				name = "SBB "
			}
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters(name, segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//SUB immediate from accumulator
		case 0b00101100:
			fallthrough
		case 0b00101101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("SUB ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//SBB immediate from accumulator
		case 0b00011100:
			fallthrough
		case 0b00011101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("SBB ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//INC register
		case 0b01000000:
			fallthrough
		case 0b01000001:
			fallthrough
		case 0b01000010:
			fallthrough
		case 0b01000011:
			fallthrough
		case 0b01000100:
			fallthrough
		case 0b01000101:
			fallthrough
		case 0b01000110:
			fallthrough
		case 0b01000111:
			builder.WriteString("INC ")
			builder.WriteString(registers[data[position]&Shared.RMMask|0b00001000])
		//DEC register
		case 0b01001000:
			fallthrough
		case 0b01001001:
			fallthrough
		case 0b01001010:
			fallthrough
		case 0b01001011:
			fallthrough
		case 0b01001100:
			fallthrough
		case 0b01001101:
			fallthrough
		case 0b01001110:
			fallthrough
		case 0b01001111:
			builder.WriteString("DEC ")
			builder.WriteString(registers[data[position]&0b00001111])

		//TEST/NOT/NEG/MUL/IMUL/DIV/IDIV R/M
		case 0b11110110:
			fallthrough
		case 0b11110111:
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			reg := data[position] & Shared.RegMask >> 3
			if reg == 1 {
				return "", newInvalidParameterErrorInvalidInstruction(position)
			}
			var name = [8]string{"TEST ", "", "NOT ", "NEG ", "MUL ", "IMUL ", "DIV ", "IDIV "}[reg]
			assembly, err = disassembleStandardParameters(name, segmentOverride, reg != 0, false, false, reg == 0, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//CMP register to R/M
		case 0b00111000:
			fallthrough
		case 0b00111001:
			fallthrough
		case 0b00111010:
			fallthrough
		case 0b00111011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("CMP ", segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//CMP immediate with accumulator
		case 0b00111100:
			fallthrough
		case 0b00111101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("CMP ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//AAM/AAD
		case 0b11010100:
			fallthrough
		case 0b11010101:
			var name string
			if data[position]&Shared.WideMask == 0 {
				name = "AAM\n"
			} else {
				name = "AAD\n"
			}
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			if data[position] != 0b00001010 {
				return "", newInvalidParameterError(position, "")
			}
			builder.WriteString(name)

		//ROL/ROR/RCL/RCR/SHL/SAL/SHR/SAR
		case 0b11010000:
			fallthrough
		case 0b11010001:
			fallthrough
		case 0b11010010:
			fallthrough
		case 0b11010011:
			countInCL := data[position]&Shared.DirectionMask != 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			reg := data[position] & Shared.RegMask >> 3
			if reg == 0b110 {
				return "", newInvalidParameterErrorInvalidInstruction(position)
			}
			var name = [8]string{"ROL ", "ROR ", "RCL ", "RCR ", "SHL ", "SHR ", "", "SAR "}[reg]
			assembly, err = disassembleStandardParameters(name, segmentOverride, true, false, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
			if countInCL {
				builder.WriteString(", CL\n")
			} else {
				builder.WriteString(", 1\n")
			}

		//AND register with R/M
		case 0b00100000:
			fallthrough
		case 0b00100001:
			fallthrough
		case 0b00100010:
			fallthrough
		case 0b00100011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("AND ", segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//AND immediate to accumulator
		case 0b00100100:
			fallthrough
		case 0b00100101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("AND ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//OR register with R/M
		case 0b00001000:
			fallthrough
		case 0b00001001:
			fallthrough
		case 0b00001010:
			fallthrough
		case 0b00001011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("OR ", segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//OR immediate to accumulator
		case 0b00001100:
			fallthrough
		case 0b00001101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("OR ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//XOR register with R/M
		case 0b00110000:
			fallthrough
		case 0b00110001:
			fallthrough
		case 0b00110010:
			fallthrough
		case 0b00110011:
			sourceInReg := data[position]&Shared.DirectionMask == 0
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("XOR ", segmentOverride, false, sourceInReg, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//XOR immediate to accumulator
		case 0b00110100:
			fallthrough
		case 0b00110101:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("XOR ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//TEST register and R/S
		case 0b10000100:
			fallthrough
		case 0b10000101:
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			assembly, err = disassembleStandardParameters("TEST ", segmentOverride, false, true, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)
		//TEST immediate with accumulator
		case 0b10101000:
			fallthrough
		case 0b10101001:
			wide := data[position]&Shared.WideMask != 0
			assembly, err = disassembleImmediateToAccumulator("TEST ", wide, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//REP
		case 0b11110010:
			fallthrough
		case 0b11110011:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("REPNZ ")
			} else {
				builder.WriteString("REPZ ")
			}
		//MOVS
		case 0b10100100:
			fallthrough
		case 0b10100101:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("MOVSB\n")
			} else {
				builder.WriteString("MOVSW\n")
			}
		//CMPS
		case 0b10100110:
			fallthrough
		case 0b10100111:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("CMPSB\n")
			} else {
				builder.WriteString("CMPSW\n")
			}
		//SCAS
		case 0b10101110:
			fallthrough
		case 0b10101111:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("SCASB\n")
			} else {
				builder.WriteString("SCASW\n")
			}
		//LODS
		case 0b10101100:
			fallthrough
		case 0b10101101:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("LODSB\n")
			} else {
				builder.WriteString("LODSW\n")
			}
		//STOS
		case 0b10101010:
			fallthrough
		case 0b10101011:
			if data[position]&Shared.WideMask == 0 {
				builder.WriteString("STOSB\n")
			} else {
				builder.WriteString("STOSW\n")
			}

		//RET intra segment with immediate
		case 0b11000010:
			fallthrough
		//RET inter segment with immediate
		case 0b11001010:
			fallthrough
		//CALL direct intra segment
		case 0b11101000:
			fallthrough
		//JMP direct intra segment
		case 0b11101001:
			instr := data[position] & 0b00001111
			var name = []string{0b0010: "RET ", 0b1000: "CALL ", 0b1001: "JMP ", 0b1010: "RETF "}[instr]
			position += 2
			if position >= dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			value := uint64(data[position-1]) | uint64(data[position])<<8
			if instr == 0b1000 || instr == 0b1001 {
				value += uint64(position + 1)
			}
			builder.WriteString(name)
			builder.WriteString(strconv.FormatUint(value, 10))
		//CALL direct inter segment
		case 0b10011010:
			fallthrough
		//JMP direct inter segment
		case 0b11101010:
			var name string
			if data[position]&0b10000 == 0 {
				name = "JMP "
			} else {
				name = "CALL "
			}
			position += 4
			if position >= dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			builder.WriteString(name)
			builder.WriteString(strconv.FormatUint(uint64(data[position-1])|uint64(data[position])<<8, 10))
			builder.WriteString(":")
			builder.WriteString(strconv.FormatUint(uint64(data[position-3])|uint64(data[position-2])<<8, 10))

		//INC/DEC/CALL/JMP/CALL far/JMP far/PUSH R/M
		case 0b11111110:
			fallthrough
		case 0b11111111:
			wide := Shared.IsolateAndShiftWide(data[position])
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			reg := data[position] & Shared.RegMask >> 3
			if reg == 0b111 || (wide == 0 && reg&0b110 != 0) {
				return "", newInvalidParameterErrorInvalidInstruction(position)
			}
			mod := data[position] & Shared.ModMask
			var name = [8]string{"INC ", "DEC ", "CALL ", "CALL ", "JMP ", "JMP ", "PUSH ", ""}[reg]
			assembly, err = disassembleStandardParameters(name, segmentOverride, true, false, false, false, false, wide, data, &position)
			if err != nil {
				return "", err
			}
			switch reg {
			case 0b010:
				fallthrough
			case 0b100:
				assembly = strings.Replace(assembly, "word", "", 1)
			case 0b011:
				fallthrough
			case 0b101:
				if mod != Shared.RegisterMode {
					assembly = strings.Replace(assembly, "word", "far", 1)
				} else {
					assembly = strings.Replace(assembly, "word", "", 1)
				}
			}
			if (reg == 0b011 || reg == 0b101) && mod == Shared.RegisterMode {
				assembly = strings.Replace(assembly, "word", "far", 1)
			}
			builder.WriteString(assembly)

		//JO
		case 0b01110000:
			fallthrough
		//JNO
		case 0b01110001:
			fallthrough
		//JB
		case 0b01110010:
			fallthrough
		//JAE
		case 0b01110011:
			fallthrough
		//JE
		case 0b01110100:
			fallthrough
		//JNE
		case 0b01110101:
			fallthrough
		//JBE
		case 0b01110110:
			fallthrough
		//JA
		case 0b01110111:
			fallthrough
		//JS
		case 0b01111000:
			fallthrough
		//JNS
		case 0b01111001:
			fallthrough
		//JP
		case 0b01111010:
			fallthrough
		//JPO
		case 0b01111011:
			fallthrough
		//JL
		case 0b01111100:
			fallthrough
		//JGE
		case 0b01111101:
			fallthrough
		//JLE
		case 0b01111110:
			fallthrough
		//JG
		case 0b01111111:
			fallthrough
		//LOOPNE
		case 0b11100000:
			fallthrough
		//LOOPE
		case 0b11100001:
			fallthrough
		//LOOP
		case 0b11100010:
			fallthrough
		//JCXZ
		case 0b11100011:
			fallthrough
		//JMP direct intra segment short
		case 0b11101011:
			var name = [255]string{
				0b01110000: "JO ",
				0b01110001: "JNO ",
				0b01110010: "JB ",
				0b01110011: "JAE ",
				0b01110100: "JE ",
				0b01110101: "JNE ",
				0b01110110: "JBE ",
				0b01110111: "JA ",
				0b01111000: "JS ",
				0b01111001: "JNS ",
				0b01111010: "JP ",
				0b01111011: "JPO ",
				0b01111100: "JL ",
				0b01111101: "JGE ",
				0b01111110: "JLE ",
				0b01111111: "JG ",
				0b11100000: "LOOPNE ",
				0b11100001: "LOOPE ",
				0b11100010: "LOOP ",
				0b11100011: "JCXZ ",
				0b11101011: "JMP ",
			}[data[position]]
			assembly, err = readAndDisassembleImmediateIntraSegmentJump(name, false, data, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString(assembly)

		//INT
		case 0b11001101:
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			builder.WriteString("INT ")
			builder.WriteString(strconv.FormatUint(uint64(data[position]), 10))

		//ESC
		case 0b11011000:
			fallthrough
		case 0b11011001:
			fallthrough
		case 0b11011010:
			fallthrough
		case 0b11011011:
			fallthrough
		case 0b11011100:
			fallthrough
		case 0b11011101:
			fallthrough
		case 0b11011110:
			fallthrough
		case 0b11011111:
			opcode := data[position] << 3
			position++
			if position == dataLength {
				return "", newInvalidParameterErrorPrematureEndOfStream(position)
			}
			var second string
			second, err = disassembleRegisterOrMemory(segmentOverride, Shared.WIDE, data, &position)
			builder.WriteString("ESC ")
			builder.WriteString(strconv.FormatUint(uint64(opcode|data[position]&Shared.RegMask>>3), 10))
			builder.WriteString(", ")
			builder.WriteString(second)

		//SEGMENT override
		case 0b00100110:
			fallthrough
		case 0b00101110:
			fallthrough
		case 0b00110110:
			fallthrough
		case 0b00111110:
			segmentOverride = segmentRegisters[data[position]&0b00011000>>3] + ":"
			continue

		//DAA
		case 0b00100111:
			fallthrough
		//DAS
		case 0b00101111:
			fallthrough
		//AAA
		case 0b00110111:
			fallthrough
		//AAS
		case 0b00111111:
			fallthrough
		//CBW
		case 0b10011000:
			fallthrough
		//CWD
		case 0b10011001:
			fallthrough
		//WAIT
		case 0b10011011:
			fallthrough
		//PUSHF
		case 0b10011100:
			fallthrough
		//POPF
		case 0b10011101:
			fallthrough
		//SAHF
		case 0b10011110:
			fallthrough
		//LAHF
		case 0b10011111:
			fallthrough
		//RET
		case 0b11000011:
			fallthrough
		//RET
		case 0b11001011:
			fallthrough
		//INT 3
		case 0b11001100:
			fallthrough
		//INTO
		case 0b11001110:
			fallthrough
		//IRET
		case 0b11001111:
			fallthrough
		//XLAT
		case 0b11010111:
			fallthrough
		//LOCK
		case 0b11110000:
			fallthrough
		//HLT
		case 0b11110100:
			fallthrough
		//CMC
		case 0b11110101:
			fallthrough
		//CLC
		case 0b11111000:
			fallthrough
		//STC
		case 0b11111001:
			fallthrough
		//CLI
		case 0b11111010:
			fallthrough
		//STI
		case 0b11111011:
			fallthrough
		//CLD
		case 0b11111100:
			fallthrough
		//STD
		case 0b11111101:
			builder.WriteString(directMappedInstructions[data[position]])
			if data[position] == 0b11110000 {
				continue
			}

		default:
			return "", &DisassembleError{Message: "invalid instruction", Pos: position}
		}

		builder.WriteString(" ; ")
		builder.WriteString(strconv.Itoa(position - startOfInsctruction))
		builder.WriteString("bytes\n")

		segmentOverride = ""
		startOfInsctruction = position
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}
