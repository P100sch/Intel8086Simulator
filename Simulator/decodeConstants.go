package Simulator

// bit mask
//goland:noinspection GoUnused
const (
  ModMask       = 0b11000000
  RMMask        = 0b00000111
  RegMask       = 0b00111000
  SegMask       = 0b00011000
  WideMask      = 0b00000001
  DirectionMask = 0b00000010
)

// parameter mode
const (
  MemoryMode   = 0b00000000
  Memory8Mode  = 0b01000000
  Memory16Mode = 0b10000000
  RegisterMode = 0b11000000
)

const _WIDE = 0b00001000

var registers = [16]string{
  "AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH",
  "AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI",
}

func Registers() [16]string {
  return registers
}

var memoryRegisters = [8]string{
  "BX + SI", "BX + DI", "BP + SI", "BP + DI", "SI", "DI", "BP", "BX",
}

func MemoryRegisters() [8]string {
  return memoryRegisters
}

var segmentRegisters = [4]string{
  "ES", "CS", "SS", "DS",
}

func SegmentRegisters() [4]string {
  return segmentRegisters
}

//goland:noinspection SpellCheckingInspection
var directMappedInstructions = [255]string{
  0b00100111: "DAA\n",
  0b00101111: "DAS\n",
  0b00110111: "AAA\n",
  0b00111111: "AAS\n",
  0b10011000: "CBW\n",
  0b10011001: "CWD\n",
  0b10011011: "WAIT\n",
  0b10011100: "PUSHF\n",
  0b10011101: "POPF\n",
  0b10011110: "SAHF\n",
  0b10011111: "LAHF\n",
  0b11000011: "RET\n",
  0b11001011: "RETF\n",
  0b11001100: "INT3\n",
  0b11001110: "INTO\n",
  0b11001111: "IRET\n",
  0b11010111: "XLAT\n",
  0b11110000: "LOCK ",
  0b11110100: "HLT\n",
  0b11110101: "CMC\n",
  0b11111000: "CLC\n",
  0b11111001: "STC\n",
  0b11111010: "CLI\n",
  0b11111011: "STI\n",
  0b11111100: "CLD\n",
  0b11111101: "STD\n",
}
