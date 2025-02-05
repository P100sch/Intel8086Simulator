package Simulator

import (
	"strconv"
	"strings"
)

type DecodeError struct {
	Message string
	Pos     int
}

func newInvalidParameterError(position int) *DecodeError {
	return &DecodeError{Message: "invalid parameters", Pos: position}
}

func (e *DecodeError) Error() string {
	return "Position " + strconv.Itoa(e.Pos) + ": " + e.Message
}

func Decode(fileData []byte) (string, error) {
	builder := strings.Builder{}

	dataLength := len(fileData)
	for position := 0; position < dataLength; position++ {
		switch {
		case fileData[position]&0b10110000 == 0b10110000:
			builder.WriteString("mov ")
			builder.WriteString(registers[fileData[position]&0b00001111])
			builder.WriteString(", ")
			position++
			if position == dataLength {
				return "", newInvalidParameterError(position)
			}
			data := uint16(fileData[position])
			if fileData[position-1]&0b00001000 == 0b00001000 {
				position++
				if position == dataLength {
					return "", newInvalidParameterError(position)
				}
				data |= uint16(fileData[position]) << 8
			}
			builder.WriteString(strconv.FormatUint(uint64(data), 10))
			builder.WriteString("\n")
		case fileData[position]&0b11111100 == 0b10001000:
			sourceFirst := fileData[position] & 0b00000010
			wide := fileData[position] & 0b00000001 << 3
			position++
			if position == dataLength {
				return "", newInvalidParameterError(position)
			}
			first := registers[wide|fileData[position]&0b00111000>>3]
			second, err := decodeRegisterMemory(wide, fileData, &position)
			if err != nil {
				return "", err
			}
			builder.WriteString("mov ")
			builder.WriteString(order(sourceFirst, first, second))
			builder.WriteString("\n")
		case fileData[position]&0b11111110 == 0b11000110:
			wide := fileData[position] & 0b00000001 << 3
			position++
			if position == dataLength {
				return "", newInvalidParameterError(position)
			}
			mod := fileData[position] & 0b11000000
			second, err := decodeRegisterMemory(wide, fileData, &position)
			if err != nil {
				return "", err
			}
			var data uint64
			position++
			if position == dataLength {
				return "", newInvalidParameterError(position)
			}
			data = uint64(fileData[position])
			if wide != 0 {
				position++
				if position == dataLength {
					return "", newInvalidParameterError(position)
				}
				data |= uint64(fileData[position]) << 8
			}
			builder.WriteString("mov ")
			builder.WriteString(second)
			builder.WriteString(", ")
			if wide == 0 && mod != 0b11000000 {
				builder.WriteString("byte ")
			} else if mod != 0b11000000 {
				builder.WriteString("word ")
			}
			builder.WriteString(strconv.FormatUint(data, 10))
			builder.WriteString("\n")
		case fileData[position]&0b11111110 == 0b10100000:
			var first string
			if fileData[position]&0b00000001 == 0 {
				first = "AL"
			} else {
				first = "AX"
			}
			position += 2
			if position >= dataLength {
				return "", newInvalidParameterError(position)
			}
			builder.WriteString("mov ")
			builder.WriteString(first)
			builder.WriteString(", [")
			builder.WriteString(strconv.FormatUint(uint64(fileData[position-1])|uint64(fileData[position])<<8, 10))
			builder.WriteString("]\n")
		case fileData[position]&0b11111110 == 0b10100010:
			var first string
			if fileData[position]&0b00000001 == 0 {
				first = "AL"
			} else {
				first = "AX"
			}
			position += 2
			if position >= dataLength {
				return "", newInvalidParameterError(position)
			}
			builder.WriteString("mov [")
			builder.WriteString(strconv.FormatUint(uint64(fileData[position-1])|uint64(fileData[position])<<8, 10))
			builder.WriteString("], ")
			builder.WriteString(first)
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

var registers = [16]string{
	"AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH",
	"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI",
}

var memoryRegisters = [8]string{
	"BX + SI", "BX + DI", "BP + SI", "BP + DI", "SI", "DI", "BP", "BX",
}

func order(sourceFirst byte, first string, second string) string {
	if sourceFirst == 0 {
		return second + ", " + first
	} else {
		return first + ", " + second
	}
}

func decodeRegisterMemory(wide byte, data []byte, position *int) (string, error) {
	var second string
	dataLength := len(data)

	mod := data[*position] & 0b11000000
	if mod == 0b11000000 {
		second = registers[wide|data[*position]&0b00000111]
	} else {
		second = memoryRegisters[data[*position]&0b00000111]

		directAddress := mod == 0 && second == "BP"
		var displacement byte
		if mod != 0 || directAddress {
			*position++
			if *position == dataLength {
				return "", newInvalidParameterError(*position)
			}
			displacement = data[*position]
			if mod == 0b01000000 {
				second += " + " + strconv.Itoa(int(int8(displacement)))
			}
		}
		if mod == 0b10000000 || directAddress {
			*position++
			if *position == dataLength {
				return "", newInvalidParameterError(*position)
			}
			if mod == 0b10000000 {
				second += " + " + strconv.Itoa(int(int16(displacement)|int16(data[*position])<<8))
			} else {
				second = strconv.FormatUint(uint64(displacement)|uint64(data[*position])<<8, 10)
			}
		}

		second = "[" + second + "]"
	}
	return second, nil
}
