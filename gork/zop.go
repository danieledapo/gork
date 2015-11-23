package gork

import (
	"fmt"
)

const (
	// optypes
	LARGE_CONSTANT    = byte(0x00)
	SMALL_CONSTANT    = byte(0X01)
	VARIABLE_CONSTANT = byte(0X02)
	OMMITTED_CONSTANT = byte(0X03)
)

const (
	ZEROOP = byte(0x00)
	ONEOP  = byte(0x01)
	TWOOP  = byte(0x02)
	VAROP  = byte(0x03)
	// v3 ignore EXTENDED OPERAND
)

type ZOp struct {
	opcode   byte
	class    byte
	optypes  []byte
	operands []uint16 // actually not all operands are large constants

	// optional values are not read,
	// the actual implementation of the functions will read them
	// possible values are:
	// - store  	byte
	// - branch(es) word
	// - text   	zstring
}

func NewZOp(seq *ZMemorySequential) *ZOp {
	zop := new(ZOp)

	opcode := seq.ReadByte()

	if opcode < 0x80 {
		zop.class = TWOOP
	} else if opcode < 0xB0 {
		zop.class = ONEOP
	} else if opcode < 0xC0 {
		zop.class = ZEROOP
	} else {
		zop.class = VAROP
	}

	switch opcode >> 6 {
	case 0x03:
		zop.configureVar(seq, opcode)
	case 0x02:
		zop.configureShort(seq, opcode)
	default:
		zop.configureLong(seq, opcode)
		// v3 ignore EXTENDED
	}

	return zop
}

func (zop *ZOp) configureVar(seq *ZMemorySequential, op byte) {
	// opcode is stored in the bottom 5 bits
	zop.opcode = op & 0x1F

	// types are stored in an additional byte
	// 2 bits per type
	// bits #7 #6 are first operand's type
	// bits #1 #0 are last operand's type
	types := seq.ReadByte()

	mask := byte(0xC0)
	for ; mask > 0; mask = mask >> 2 {
		ty := types & mask
		if ty == OMMITTED_CONSTANT {
			break
		}
		zop.optypes = append(zop.optypes, ty)
		zop.operands = append(zop.operands, readOpType(seq, ty))
	}

	for ; mask > 0; mask = mask >> 2 {
		if types&mask != OMMITTED_CONSTANT {
			panic("non omitted type after omitted one!")
		}
	}

	if zop.class == TWOOP && len(zop.optypes) != 2 {
		panic("2op in var form does not have 2 ops")
	}
}

func (zop *ZOp) configureShort(seq *ZMemorySequential, op byte) {
	// opcode is stored in the bottom 4 bits
	zop.opcode = op & 0x0F

	if zop.class == ONEOP {
		zop.operands = make([]uint16, 1)
		zop.optypes = make([]byte, 1)

		// optype is stored in bits #4 #5
		zop.optypes[0] = op & 0x18
		zop.operands[0] = readOpType(seq, zop.optypes[0])
	} // ignore ZEROOP
}

func (zop *ZOp) configureLong(seq *ZMemorySequential, op byte) {
	// always 2OP
	// opcode is stored in the bottom 5 bits
	zop.opcode = op & 0x1F

	zop.operands = make([]uint16, 2)
	zop.optypes = make([]byte, 2)

	// the type of operand #1 is in bit #6
	// the type of operand #2 is in bit #5
	// if bit == 0 then type is SMALL_CONSTANT,
	// 		otherwise it is VARIABLE_CONSTANT
	for i := uint8(0); i < 2; i++ {
		if zop.opcode&(0x20>>i) == 0x00 {
			zop.optypes[i] = SMALL_CONSTANT
		} else {
			zop.optypes[i] = VARIABLE_CONSTANT
		}
		zop.operands[i] = readOpType(seq, zop.optypes[i])
	}
}

func readOpType(seq *ZMemorySequential, optype byte) uint16 {
	if optype == LARGE_CONSTANT {
		return seq.ReadWord()
	} else {
		return uint16(seq.ReadByte())
	}
}

func (zop *ZOp) String() string {
	// not properly formatted

	ret := ""

	ret += fmt.Sprintf("Opcode: %d ", zop.opcode)

	switch zop.class {
	case ZEROOP:
		ret += "0OP"
	case ONEOP:
		ret += "1OP"
	case TWOOP:
		ret += "2OP"
	case VAROP:
		ret += "VAR"
	}
	ret += "\n"

	ret += fmt.Sprintln("Operands:\n")

	for i := range zop.operands {
		ty := zop.optypes[i]
		operand := zop.operands[i]

		ret += fmt.Sprintf("  %X ", operand)

		switch ty {
		case LARGE_CONSTANT:
			ret += "LARGE_CONSTANT"
		case SMALL_CONSTANT:
			ret += "SMALL_CONSTANT"
		case VARIABLE_CONSTANT:
			ret += "VARIABLE_CONSTANT"
		case OMMITTED_CONSTANT:
			ret += "OMMITTED_CONSTANT"
		}
		ret += "\n"
	}

	return ret
}
