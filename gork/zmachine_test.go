package gork

import "testing"

var someRoutines []*ZRoutine = []*ZRoutine{
	&ZRoutine{
		addr:    0,
		retAddr: 0,
		locals:  []uint16{},
	},
	&ZRoutine{
		addr:    42,
		retAddr: 73,
		locals:  []uint16{42, 73},
	},
	nil,
}

func TestZStackPush(t *testing.T) {
	stack := ZStack{}

	for _, routine := range someRoutines {
		stack.Push(routine)

		if stack[len(stack)-1] != routine {
			t.Fail()
		}
	}
}

func TestZStackPop(t *testing.T) {
	stack := ZStack(someRoutines[:])

	i := len(someRoutines) - 1

	for len(stack) > 0 {
		top := stack.Top()
		popped := stack.Pop()

		if top != popped || popped != someRoutines[i] {
			t.Fail()
		}
		i--
	}

}
