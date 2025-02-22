package Simulation

import (
	"fmt"
	"slices"
	"strings"
)

// end exclusive
func readInstruction(start, end uint16) []byte {
	baseAddress := int(CS) << 4
	if start > end {
		return slices.Concat(Memory[baseAddress+int(start):], Memory[:baseAddress+int(end)+1])
	} else {
		return Memory[baseAddress+int(start) : baseAddress+int(end)+1]
	}
}

// formatState formats the state in a loggable format.
func formatState() string {
	builder := strings.Builder{}
	if TF != 0 {
		builder.WriteString("T")
	} else {
		builder.WriteByte(' ')
	}
	if DF != 0 {
		builder.WriteString("D")
	} else {
		builder.WriteByte(' ')
	}
	if IF != 0 {
		builder.WriteString("I")
	} else {
		builder.WriteByte(' ')
	}
	if OF != 0 {
		builder.WriteString("O")
	} else {
		builder.WriteByte(' ')
	}
	if SF != 0 {
		builder.WriteString("S")
	} else {
		builder.WriteByte(' ')
	}
	if ZF != 0 {
		builder.WriteString("Z")
	} else {
		builder.WriteByte(' ')
	}
	if AF != 0 {
		builder.WriteString("A")
	} else {
		builder.WriteByte(' ')
	}
	if PF != 0 {
		builder.WriteString("P")
	} else {
		builder.WriteByte(' ')
	}
	if CF != 0 {
		builder.WriteString("C")
	} else {
		builder.WriteByte(' ')
	}
	return fmt.Sprintf("AX:0x%04x BX:0x%04x CX:0x%04x DX:0x%04x SP:0x%04x BP:0x%04x SI:0x%04x DI:0x%04x IP:0x%04x CS:0x%04x DS:0x%04x SS:0x%04x ES:0x%04x F:%9s", AX, BX, CX, DX, SP, BP, SI, DI, IP, CS, DS, SS, ES, builder.String())
}
