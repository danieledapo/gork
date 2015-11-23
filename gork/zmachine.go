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
	// pc is mem.pos
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

	return &ZMachine{header: header, seq: mem.GetSequential(header.pc), quitted: false, objects: objects}
}

func (zm *ZMachine) GetVarAt(varnum uint16) uint16 {
	if varnum == 0 {
		// top of stack
		return zm.stack.Top().locals[len(zm.stack.Top().locals)-1]
	} else if varnum < 0x10 {
		// local variable
		return zm.stack.Top().locals[varnum-1]
	} else {
		// global variable
		return zm.seq.mem.WordAt(zm.header.globalsPos + (varnum-0x10)*2)
	}
}

func (zm *ZMachine) StoreVarAt(varnum uint16, val uint16) {
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
		globalAddr := zm.header.globalsPos + (varnum-0x10)*2
		zm.seq.mem.WriteWordAt(globalAddr, val)
	}
}

func (zm *ZMachine) StoreReturn(val uint16) {
	varnum := zm.seq.ReadByte()
	zm.StoreVarAt(uint16(varnum), val)
}

func (zm *ZMachine) InterpretAll() {
	for !zm.quitted {
		zm.Interpret()
	}
}

func (zm *ZMachine) Interpret() {
	op := NewZOp(zm.seq)
	fmt.Printf("instruction %d class: %d\nPC: %X\n", op.opcode, op.class, zm.seq.pos)

	switch op.class {
	case ZEROOP:
		zeroOpFuncs[op.opcode](zm)
	case ONEOP:
		oneOpFuncs[op.opcode](zm, op.operands[0])
	case TWOOP:
		twoOpFuncs[op.opcode](zm, op.operands[0], op.operands[1])
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
