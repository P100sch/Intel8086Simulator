package Simulator

import (
	"strconv"
	"strings"
)

type DecodeError struct {
	Message string
	Pos     int
}

func (e *DecodeError) Error() string {
	return "Position " + strconv.Itoa(e.Pos) + ": " + e.Message
}

func Decode(fileData []byte) (string, error) {
	builder := strings.Builder{}

	for position := 0; position < len(fileData); position++ {
		switch {
		case fileData[position]&0b11111100 == 0b10001000:
			builder.WriteString("mov ")
			position++
			if position > len(fileData) {
				return "", &DecodeError{Message: "invalid parameters", Pos: position}
			}
			if fileData[position]&0b11000000 != 0b11000000 {
				return "", &DecodeError{Message: "only register to register move supported", Pos: position}
			}
			builder.WriteString(decodeRegisters(fileData[position-1], fileData[position]))
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

var registers = [16]string{
	"AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH",
	"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI",
}

func decodeRegisters(instruction byte, parameters byte) string {
	var reg1, reg2 string
	wideBit := (instruction & 0b00000001) << 3
	reg1 = registers[wideBit|(parameters&0b00111000)>>3]
	reg2 = registers[wideBit|(parameters&0b00000111)]
	if instruction&0b00000010 == 1 {
		return reg1 + ", " + reg2
	} else {
		return reg2 + ", " + reg1
	}
}
