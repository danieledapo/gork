package gork

import (
	"fmt"
)

// bottom is in #0
// top is in #len(stack-1)
type ZStack []*ZRoutine

func (zstack *ZStack) Push(routine *ZRoutine) {
	*zstack = append(*zstack, routine)
}

func (zstack *ZStack) Pop() *ZRoutine {
	last := len(*zstack) - 1
	ret := (*zstack)[last]
	*zstack = (*zstack)[:last]
	return ret
}

func (zstack *ZStack) Top() *ZRoutine {
	return (*zstack)[len(*zstack)-1]
}

type ZMachine struct {
	header *ZHeader
	// pc is seq.pos
	seq     *ZMemorySequential
	objects []*ZObject
	stack   ZStack
	quitted bool
}

func NewZMachine(mem *ZMemory, header *ZHeader) *ZMachine {
	// cache objects
	count := ZObjectsCount(mem, header)
	objects := make([]*ZObject, count)

	for i := uint8(1); i <= count; i++ {
		// objects are 1-based
		objects[i-1] = NewZObject(mem, i, header)
	}

	stack := ZStack{}
	stack.Push(MainRoutine(mem, header))

	return &ZMachine{
		header:  header,
		seq:     mem.GetSequential(header.pc),
		objects: objects,
		quitted: false,
		stack:   stack,
	}
}

func (zm *ZMachine) GetVarAt(varnum byte) uint16 {
	if varnum == 0 {
		// top of stack
		return zm.stack.Top().locals[len(zm.stack.Top().locals)-1]
	} else if varnum < 0x10 {
		// local variable
		return zm.stack.Top().locals[varnum-1]
	} else {
		// global variable
		return zm.seq.mem.WordAt(zm.header.globalsPos + (uint16(varnum)-0x10)*2)
	}
}

func (zm *ZMachine) StoreVarAt(varnum byte, val uint16) {
	if varnum == 0 {
		// push to top of the stack
		topRoutinelocals := &zm.stack.Top().locals
		*topRoutinelocals = append(*topRoutinelocals, val)
	} else if varnum < 0x10 {
		// local variable
		// starting from 0
		zm.stack.Top().locals[varnum-1] = val
	} else {
		// global variable
		// globals table is a table of 240 words
		globalAddr := zm.header.globalsPos + uint16(varnum-0x10)*2
		zm.seq.mem.WriteWordAt(globalAddr, val)
	}
}

func (zm *ZMachine) StoreReturn(val uint16) {
	varnum := zm.seq.ReadByte()
	zm.StoreVarAt(varnum, val)
}

func (zm *ZMachine) Branch(conditionOk bool) {
	info := zm.seq.ReadByte()

	// if bit #7 is set than branch on true
	branchOnTrue := (info >> 7) != 0x00

	// offset is relative to current PC and it can be negative
	var offset int16

	// if bit #6 is set than the offset is stored in the bottom
	// 6 bits
	if info&0x40 != 0x00 {
		offset = int16(info & 0x3F)
	} else {
		// if bit #6 is clear than the offset is store in a 14 bit signed
		// integer composed by the bottom 5 bits of info and 8 bits of an
		// additional byte
		firstPart := info & 0x3F

		// if sign bit(#6) is set then it's a negative number
		// in two complement form, so set the bits #6 and #7 too
		if firstPart&0x20 != 0x00 {
			firstPart |= 0x3 << 6
		}

		offset = int16(int16(firstPart<<8) | int16(zm.seq.ReadByte()))
	}

	// jump if conditionOk and branchOnTrue are both true or false
	if conditionOk == branchOnTrue {
		if offset == 0 {
			// offset of 0 means return false from current routine
			ZReturnFalse(zm)
		} else if offset == 1 {
			// offset of 1 means return false from current routine
			ZReturnTrue(zm)
		} else {
			// otherwise we move to instruction to the given offset
			zm.seq.pos = zm.CalcJumpAddress(int32(offset))
			// fmt.Printf("Jumping to 0x%X offset %d\n", jumpAddr, offset)
		}
	}
}

func (zm *ZMachine) CalcJumpAddress(offset int32) uint16 {
	// Address after branch data + Offset - 2
	return uint16(int32(zm.seq.pos) + int32(offset) - 2)
}

func (zm *ZMachine) InterpretAll() {
	for !zm.quitted {
		zm.Interpret()
	}
}

func (zm *ZMachine) Interpret() {
	op := NewZOp(zm)
	fmt.Printf("instruction %2d class: %d PC: %X\n", op.opcode, op.class, zm.seq.pos)

	switch op.class {
	case ZEROOP:
		zeroOpFuncs[op.opcode](zm)
	case ONEOP:
		oneOpFuncs[op.opcode](zm, op.operands[0])
	case TWOOP:
		if op.opcode == 1 {
			// ZJe is a two op func but it accepts VAR count of args,
			// so we must handle separetly
			ZJe(zm, op.operands)
		} else {
			twoOpFuncs[op.opcode](zm, op.operands[0], op.operands[1])
		}
	case VAROP:
		varOpFuncs[op.opcode](zm, op.operands)
	}
}

func (zm *ZMachine) String() string {
	// not properly formatted
	ret := ""

	ret += fmt.Sprintf("PC: %X\n", zm.seq.pos)
	ret += fmt.Sprintf("Stack: %s\n", zm.stack)
	ret += fmt.Sprintf("Quitted: %b\n", zm.quitted)

	return ret
}

func (zstack *ZStack) String() string {
	// not properly formatted
	ret := ""

	ret += fmt.Sprintf("Size: %d\n", len(*zstack))
	ret += fmt.Sprintf("Current routine at %X\n", zstack.Top().addr)

	return ret
}
