package Simulation

import (
  "fmt"
  "strings"
)

const (
  _L = 0b0000000011111111
  _H = 0b1111111100000000
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

func state() string {
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
