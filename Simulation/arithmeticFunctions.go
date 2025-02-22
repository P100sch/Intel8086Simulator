package Simulation

func addAndUpdateFlags(addend1, addend2 uint16, wide bool) uint16 {
	var maxValue uint32
	var signBit uint16
	if wide {
		maxValue = uint32(_W_MAX)
		signBit = _W_SIGN
	} else {
		maxValue = uint32(_B_MAX)
		signBit = _B_SIGN
	}
	sum := uint32(addend1)&maxValue + uint32(addend2)&maxValue
	truncatedSum := uint16(sum & maxValue)

	if sum > maxValue {
		CF = 1
	} else {
		CF = 0
	}
	if (addend1&0b1111)+(addend2&0b1111) > 0b1111 {
		AF = 1
	} else {
		AF = 0
	}

	sign1 := addend1 & signBit
	sign2 := addend2 & signBit
	signSum := truncatedSum & signBit
	if sign1&sign2 == signSum || sign1|sign2 == signSum {
		OF = 0
	} else {
		OF = 1
	}
	setCommonFlags(truncatedSum, signBit)
	return truncatedSum
}

func subAndUpateFlags(minuend, subtrahend uint16, wide bool) uint16 {
	var maxValue uint16
	if wide {
		maxValue = _W_MAX
	} else {
		maxValue = uint16(_B_MAX)
	}
	negatedSubtrahend := subtrahend ^ maxValue + 1
	difference := addAndUpdateFlags(minuend, negatedSubtrahend, wide)

	if minuend < subtrahend {
		CF = 1
	} else {
		CF = 0
	}
	if minuend&0b1111 < subtrahend&0b1111 {
		AF = 1
	} else {
		AF = 0
	}

	return difference
}

func wrapAdd(addend1, addend2 uint16) uint16 {
	return uint16((uint32(addend1) + uint32(addend2)) & uint32(_W_MAX))
}

func wrapIncrement(x uint16) uint16 {
	return wrapAdd(x, 1)
}
