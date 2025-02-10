package Disassembly

import (
  "strconv"

  "github.com/P100sch/Intel8086Simulator/Simulation/Shared"
)

// disassembleStandardParameters disassembles a standard parameter in data at position and returns the full assembly instruction
//   - name contains the instructions name
//   - segmentOverride contains a segment override for the register/memory portion of the parameters, if applicable
//   - noFirst is used if the register part doesn't contain a disassembleable register
//   - sourceInReg defines if the source portion is in the register or register/memory portion
//   - segmentRegister defines if register portion contains a segment register instead
//   - immediate defines if an immediate follows that's used instead of the register portion
//   - signExtended defines if the immediate should be sign extended
//   - wide byte defining register, memory and immediate width. Can be 0 or 8.
//   - data containing instruction stream
//   - position inside of instruction stream
//
// Possible cause of errors are:
//   - invalid segment register
//   - end of instruction stream reached before complete decoding
func disassembleStandardParameters(name, segmentOverride string, noFirst, sourceInReg, segmentRegister, immediate, signExtended bool, wide byte, data []byte, position *int) (string, error) {
  var err error
  var first, second string
  var mod = data[*position] & Shared.ModMask
  if !noFirst && !immediate {
    if segmentRegister {
      var valid bool
      first, valid = disassembleSegmentRegister(data[*position])
      if !valid {
        return "", newInvalidParameterError(*position, "invalid segment register")
      }
    } else {
      first = registers[wide|data[*position]&Shared.RegMask>>3]
    }
  }
  second, err = disassembleRegisterOrMemory(segmentOverride, wide, data, position)
  if err != nil {
    return "", err
  }
  if immediate {
    first, err = readData(true, wide != 0 && !signExtended, data, position)
    if mod != Shared.RegisterMode {
      if wide == 0 {
        first = "byte " + first
      } else {
        first = "word " + first
      }
    }
  }
  if noFirst {
    if mod != Shared.RegisterMode {
      if wide == 0 {
        second = "byte " + second
      } else {
        second = "word " + second
      }
    }
    return name + second, nil
  }
  return name + order(sourceInReg || immediate, first, second), nil
}

// disassembleRegisterOrMemory disassembles the register/memory portion of parameters in data at position
//   - segmentOverride contains the segment register override, if applicable
//   - wide defines if register/memory with. Can be 0 or 8
//   - data instruction stream
//   - position in instruction stream
//
// Possible errors:
//   - end of instruction stream reached before complete decoding
func disassembleRegisterOrMemory(segmentOverride string, wide byte, data []byte, position *int) (string, error) {
  var second string
  mod := data[*position] & Shared.ModMask
  if mod == Shared.RegisterMode {
    second = registers[wide|data[*position]&Shared.RMMask]
  } else {
    second = memoryRegisters[data[*position]&Shared.RMMask]

    if mod != Shared.MemoryMode || second == "BP" {
      var displacement string
      displacement, err := readData(mod != Shared.MemoryMode, mod == Shared.MemoryMode || mod == Shared.Memory16Mode, data, position)
      if err != nil {
        return "", err
      }
      if mod != Shared.MemoryMode {
        second += " + " + displacement
      } else {
        second = displacement
      }
    }

    second = segmentOverride + "[" + second + "]"
  }
  return second, nil
}

// readData reads a chunk of data as a number and strings it
//   - signed if the number is signed
//   - wide if the data is 16bits wide
//   - data instruction stream
//   - position in instruction stream
//
// Possible errors:
//   - end of instruction stream reached before complete decoding
func readData(signed bool, wide bool, data []byte, position *int) (string, error) {
  length := len(data)
  var first byte
  *position++
  if *position == length {
    return "", newInvalidParameterErrorPrematureEndOfStream(*position)
  }
  first = data[*position]
  if !wide {
    if signed {
      return strconv.Itoa(int(int8(first))), nil
    } else {
      return strconv.FormatUint(uint64(first), 10), nil
    }
  }
  *position++
  if *position == length {
    return "", newInvalidParameterErrorPrematureEndOfStream(*position)
  }
  if signed {
    return strconv.Itoa(int(int16(first) | int16(data[*position])<<8)), nil
  } else {
    return strconv.FormatUint(uint64(first)|uint64(data[*position])<<8, 10), nil
  }
}

// disassembleSegmentRegister tries to disassemble register portion of parameters as a segment register
func disassembleSegmentRegister(parameters byte) (string, bool) {
  if parameters&0b00100000 != 0 {
    return "", false
  }
  return segmentRegisters[parameters&Shared.SegMask>>3], true
}

// getAndShiftWide gets the wide bit from the instruction and shifts it up 3 bits for usage in decoding
func getAndShiftWide(instruction byte) byte {
  return instruction & Shared.WideMask << 3
}

// order concatenates first and second. Source is put last
func order(sourceInFirst bool, first string, second string) string {
  if sourceInFirst {
    return second + ", " + first
  } else {
    return first + ", " + second
  }
}

func disassembleImmediateToAccumulator(name string, wide bool, data []byte, position *int) (string, error) {
  value, err := readData(true, wide, data, position)
  if err != nil {
    return "", err
  }
  if wide {
    return name + "AX, " + value, nil
  }
  return name + "AL, " + value, nil
}

func readAndDisassembleImmediateIntraSegmentJump(name string, wide bool, data []byte, position *int) (string, error) {
  length := len(data)
  var first byte
  var value int
  *position++
  if *position == length {
    return "", newInvalidParameterErrorPrematureEndOfStream(*position)
  }
  first = data[*position]
  if !wide {
    value = int(int8(first))
  } else {
    *position++
    if *position == length {
      return "", newInvalidParameterErrorPrematureEndOfStream(*position)
    }
    value = int(int16(first) | int16(data[*position])<<8)
  }
  value += 2
  if value == 0 {
    return name + " $+0", nil
  }
  if value < 0 {
    return name + " $" + strconv.Itoa(value) + "+0", nil
  }
  return name + " $+" + strconv.Itoa(value) + "+0", nil
}
