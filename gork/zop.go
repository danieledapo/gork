package gork

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
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
	zm       *ZMachine
	opcode   byte
	class    byte
	optypes  []byte
	operands []uint16 // actually not all operands are large constants
	name     string

	// optional values are not read,
	// the actual implementation of the functions will read them
	// possible values are:
	// - store  	byte
	// - branch(es) word
	// - text   	zstring
}

func NewZOp(zm *ZMachine) *ZOp {
	zop := new(ZOp)

	zop.zm = zm

	opcode := zm.seq.ReadByte()

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
		zop.configureVar(opcode)
	case 0x02:
		zop.configureShort(opcode)
	default:
		zop.configureLong(opcode)
		// v3 ignore EXTENDED
	}

	zop.name = zop.getOpName()

	return zop
}

func (zop *ZOp) configureVar(op byte) {
	// opcode is stored in the bottom 5 bits
	zop.opcode = op & 0x1F

	// if bit #5 is 0 then it's TWOOP
	if ((op >> 5) & 0x01) == 0 {
		zop.class = TWOOP
	} else {
		zop.class = VAROP
	}

	// types are stored in an additional byte
	// 2 bits per type
	// bits #7 #6 are first operand's type
	// bits #1 #0 are last operand's type
	types := zop.zm.seq.ReadByte()

	i := 6

	for ; i >= 0; i -= 2 {
		ty := (types >> byte(i)) & 0x03
		if ty == OMMITTED_CONSTANT {
			break
		}
		zop.optypes = append(zop.optypes, ty)
		zop.operands = append(zop.operands, zop.readOpType(ty))
	}

	for ; i >= 0; i -= 2 {
		if (types>>byte(i))&0x03 != OMMITTED_CONSTANT {
			log.Fatal("non omitted type after omitted one!")
		}
	}

	// following seems reasonable, but in practice it's useless
	// because for instance ZJe is TWOOP but it actually accepts
	// 3 args
	// if zop.class == TWOOP && len(zop.optypes) != 2 {
	// 	log.Fatalf("PC: %d 2op %d in var form does not have 2 ops %v\n",
	// 		zop.zm.seq.pos, zop.opcode, zop.operands)
	// }
}

func (zop *ZOp) configureShort(op byte) {
	// opcode is stored in the bottom 4 bits
	zop.opcode = op & 0x0F

	if zop.class == ONEOP {
		zop.operands = make([]uint16, 1)
		zop.optypes = make([]byte, 1)

		// optype is stored in bits #4 #5
		zop.optypes[0] = (op >> 4) & 0x03
		zop.operands[0] = zop.readOpType(zop.optypes[0])
	} // ignore ZEROOP
}

func (zop *ZOp) configureLong(op byte) {
	// always 2OP
	// opcode is stored in the bottom 5 bits
	zop.opcode = op & 0x1F

	zop.operands = make([]uint16, 2)
	zop.optypes = make([]byte, 2)

	// the type of operand #1 is in bit #6
	// the type of operand #2 is in bit #5
	// if bit == 0 then type is SMALL_CONSTANT,
	// 		otherwise it is VARIABLE_CONSTANT
	for i := byte(0); i < 2; i++ {
		bit := op >> (6 - i) & 0x01
		if bit == 0x00 {
			zop.optypes[i] = SMALL_CONSTANT
		} else {
			zop.optypes[i] = VARIABLE_CONSTANT
		}

		zop.operands[i] = zop.readOpType(zop.optypes[i])
	}
}

func (zop *ZOp) readOpType(optype byte) uint16 {
	if optype == LARGE_CONSTANT {
		return zop.zm.seq.ReadWord()
	} else if optype == VARIABLE_CONSTANT {
		tmp := zop.zm.GetVarAt(zop.zm.seq.ReadByte())
		return tmp
	} else {
		return uint16(zop.zm.seq.ReadByte())
	}
}

func (zop *ZOp) getOpName() string {
	var fn interface{} = nil

	switch zop.class {
	case ZEROOP:
		if int(zop.opcode) < len(zeroOpFuncs) {
			fn = zeroOpFuncs[zop.opcode]
		}
	case ONEOP:
		if int(zop.opcode) < len(oneOpFuncs) {
			fn = oneOpFuncs[zop.opcode]
		}
	case TWOOP:
		if zop.opcode == 1 {
			// ZJe is a two op func but it accepts VAR count of args,
			// so we must handle separetly
			return "ZJe"
		} else if int(zop.opcode) < len(twoOpFuncs) {
			fn = twoOpFuncs[zop.opcode]
		}
	case VAROP:
		if int(zop.opcode) < len(varOpFuncs) {
			fn = varOpFuncs[zop.opcode]
		}
	}
	return getFuncName(fn, "unknown opcode name")
}

func getFuncName(fn interface{}, errFnName string) string {
	value := reflect.ValueOf(fn)

	if !value.IsValid() {
		return errFnName
	}

	completeName := runtime.FuncForPC(value.Pointer()).Name()

	if completeName == "" {
		return errFnName
	}

	// keep only function name
	return strings.Split(completeName, ".")[1]
}

func (zop *ZOp) String() string {
	// not properly formatted

	ret := ""

	ret += fmt.Sprintln("  Op:", zop.name)
	ret += fmt.Sprintf("  Opcode: %d ", zop.opcode)

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

	ret += fmt.Sprintln("  Operands:")

	for i := range zop.operands {
		ty := zop.optypes[i]
		operand := zop.operands[i]

		ret += fmt.Sprintf("    %2X ", operand)

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
