package Shared

// IsolateAndShiftWide gets the wide bit from the instruction and shifts it up 3 bits for usage in decoding
func IsolateAndShiftWide(instruction byte) byte {
  return instruction & WideMask << 3
}
