package Disassembly

var registers = [16]string{
  "AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH",
  "AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI",
}

// Registers get all valid registers
//goland:noinspection GoUnusedExportedFunction
func Registers() [16]string {
  return registers
}

var memoryRegisters = [8]string{
  "BX + SI", "BX + DI", "BP + SI", "BP + DI", "SI", "DI", "BP", "BX",
}

// MemoryRegisters get all register memory offset options
//goland:noinspection GoUnusedExportedFunction
func MemoryRegisters() [8]string {
  return memoryRegisters
}

var segmentRegisters = [4]string{
  "ES", "CS", "SS", "DS",
}

// SegmentRegisters get all valid segment registers
//goland:noinspection GoUnusedExportedFunction
func SegmentRegisters() [4]string {
  return segmentRegisters
}

//goland:noinspection SpellCheckingInspection
var directMappedInstructions = [255]string{
  0b00100111: "DAA",
  0b00101111: "DAS",
  0b00110111: "AAA",
  0b00111111: "AAS",
  0b10011000: "CBW",
  0b10011001: "CWD",
  0b10011011: "WAIT",
  0b10011100: "PUSHF",
  0b10011101: "POPF",
  0b10011110: "SAHF",
  0b10011111: "LAHF",
  0b11000011: "RET",
  0b11001011: "RETF",
  0b11001100: "INT3",
  0b11001110: "INTO",
  0b11001111: "IRET",
  0b11010111: "XLAT",
  0b11110000: "LOCK ",
  0b11110100: "HLT",
  0b11110101: "CMC",
  0b11111000: "CLC",
  0b11111001: "STC",
  0b11111010: "CLI",
  0b11111011: "STI",
  0b11111100: "CLD",
  0b11111101: "STD",
}
