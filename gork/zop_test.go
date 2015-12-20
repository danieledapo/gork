package gork

import (
	"fmt"
	"testing"
)

// TODO test VARIABLE_CONSTANT

var zopBuf [][]byte = [][]byte{
	[]byte{
		0xE0, 0x3, 0x2A, 0x39, 0x80, 0x10, 0xFF, 0xFF,
	},
	[]byte{
		0x8C, 0xFF, 0xD7,
	},
	[]byte{
		0x0D, 0x10, 0xB4,
	},
	[]byte{
		0xB2,
	},
}

var zopExpected []ZOp = []ZOp{
	ZOp{
		opcode: 0,
		class:  VAROP,
		optypes: []byte{
			LARGE_CONSTANT,
			LARGE_CONSTANT,
			LARGE_CONSTANT,
		},
		operands: []uint16{
			0x2A39, 0x8010, 0xFFFF,
		},
		name: "ZCall",
	},
	ZOp{
		opcode: 12,
		class:  ONEOP,
		optypes: []byte{
			LARGE_CONSTANT,
		},
		operands: []uint16{
			0xFFD7,
		},
		name: "ZJump",
	},
	ZOp{
		opcode: 13,
		class:  TWOOP,
		optypes: []byte{
			SMALL_CONSTANT,
			SMALL_CONSTANT,
		},
		operands: []uint16{
			0x10, 0xB4,
		},
		name: "ZStore",
	},
	ZOp{
		opcode:   2,
		class:    ZEROOP,
		optypes:  []byte{},
		operands: []uint16{},
		name:     "ZPrint",
	},
}

func TestZOP(t *testing.T) {
	for i, mem := range zopBuf {

		zmem := ZMemory(mem)
		zmachine := &ZMachine{
			header: &ZHeader{},
			seq:    zmem.GetSequential(0),
		}

		zop := NewZOp(zmachine)
		expected := zopExpected[i]

		if zop.opcode != expected.opcode || zop.class != expected.class ||
			zop.name != expected.name {
			t.Fail()
		}

		for j, ty := range zop.optypes {
			if ty != expected.optypes[j] {
				t.Fail()
			}
		}

		for j, operand := range zop.operands {
			if operand != expected.operands[j] {
				t.Fail()
			}
		}
	}
}

func TestGetOpName(t *testing.T) {
	// must not crash
	fmt.Println((&ZOp{class: TWOOP, opcode: 100}).getOpName())
	fmt.Println((&ZOp{class: ONEOP, opcode: 99}).getOpName())
	fmt.Println((&ZOp{class: TWOOP, opcode: 1}).getOpName())
	fmt.Println((&ZOp{class: TWOOP, opcode: 4}).getOpName())
	fmt.Println((&ZOp{class: ZEROOP, opcode: 99}).getOpName())
}
