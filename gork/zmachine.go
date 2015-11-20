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
	header  *ZHeader
	story   *ZStory
	objects []*ZObject
	pc      uint16
	stack   ZStack
	quitted bool
}

func NewZMachine(story *ZStory, header *ZHeader) *ZMachine {
	// cache objects
	count := ZObjectsCount(story, header)
	objects := make([]*ZObject, count)

	for i := uint8(1); i <= count; i++ {
		objects[i] = NewZObject(story, i, header)
	}

	return &ZMachine{header: header, story: story, pc: header.pc, quitted: false, objects: objects}
}

func (zm *ZMachine) StoreAt(addr uint16, val uint16) {
	if addr == 0 {
		// push to top of the stack
		topRoutinelocals := &zm.stack.Top().locals
		*topRoutinelocals = append(*topRoutinelocals, val)
	} else if addr < 0x10 {
		// local variable
		// starting from 0
		zm.stack.Top().locals[addr-1] = val
	} else {
		// global variable
		// globals table is a table of 240 words
		globalAddr := zm.header.globalsPos + (addr-0x10)*2
		zm.story.WriteWordAt(globalAddr, val)
	}
}

func (zm *ZMachine) StoreReturn(val uint16) {
	storePos := zm.story.ReadByte()
	zm.StoreAt(uint16(storePos), val)
}

func (zm *ZMachine) InterpretAll() {
	for !zm.quitted {
		zm.Interpret()
	}
}

func (zm *ZMachine) Interpret() {
	op := NewZOp(zm.story, zm.pc)

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

	ret += fmt.Sprintf("PC: %X\n", zm.pc)
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
