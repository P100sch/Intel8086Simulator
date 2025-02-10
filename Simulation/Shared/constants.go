package Shared

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
