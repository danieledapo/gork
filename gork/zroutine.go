package gork

import (
	"fmt"
)

// aka StackFrame
type ZRoutine struct {
	addr    uint16
	retAddr uint16
	locals  []uint16
}

func NewZRoutine(story *ZStory, addr uint16, retAddr uint16) *ZRoutine {
	if !IsPackedAddress(addr) {
		panic("attempt to read routine at non packed address")
	}

	story.pos = addr

	routine := new(ZRoutine)
	routine.retAddr = retAddr

	routine.addr = addr
	numLocals := story.ReadByte()

	routine.locals = make([]uint16, numLocals)

	for i := byte(0); i < numLocals; i++ {
		routine.locals[i] = story.ReadWord()
	}

	return routine
}

func MainRoutine(story *ZStory, header *ZHeader) *ZRoutine {
	return NewZRoutine(story, PackedAddress(header.pc), 0)
}

func (routine *ZRoutine) String() string {
	ret := fmt.Sprintf("Routine at %X\nLocals: [", routine.addr)

	tmp := ""
	if len(routine.locals) > 0 {
		for _, local := range routine.locals {
			tmp += fmt.Sprintf("%X, ", local)
		}
		ret += tmp[:len(tmp)-2]
	}
	ret += "]\n"

	return ret
}
