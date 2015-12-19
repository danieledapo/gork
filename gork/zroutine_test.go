package gork

import "testing"

var zroutineBuf [][]byte = [][]byte{
	[]byte{
		2,
		0x00, 0x2a,
		0x00, 0x49,
	},
	[]byte{
		0,
	},
}

var zroutineExpected []ZRoutine = []ZRoutine{
	ZRoutine{
		addr:    0,
		retAddr: 42,
		locals: []uint16{
			42, 73,
		},
	},
	ZRoutine{
		addr:    0,
		retAddr: 42,
		locals:  []uint16{},
	},
}

func TestZRoutine(t *testing.T) {
	for i, buf := range zroutineBuf {
		mem := NewZMemory(buf)

		routine := NewZRoutine(mem.GetSequential(0), 42)
		expected := zroutineExpected[i]

		if expected.addr != routine.addr ||
			expected.retAddr != routine.retAddr {
			t.Fail()
		}

		for i, local := range expected.locals {
			if local != routine.locals[i] {
				t.Fail()
			}
		}
	}
}
