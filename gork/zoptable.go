package gork

type ZeroOpFunc func(*ZMachine)
type OneOpFunc func(*ZMachine, uint16)
type TwoOpFunc func(*ZMachine, uint16, uint16)
type VarOpFunc func(*ZMachine, []uint16)

var zeroOpFuncs = []ZeroOpFunc{
	ZReturnTrue,
	ZReturnFalse,
}

var oneOpFuncs = []OneOpFunc{
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	ZReturn,
}

var twoOpFuncs = []TwoOpFunc{}

var varOpFuncs = []VarOpFunc{
	ZCall,
}

func ZCall(zm *ZMachine, operands []uint16) {
	routineAddr := PackedAddress(uint16(zm.story.ReadByte()))

	if routineAddr == 0 {
		zm.stack.Push(nil) // removed by following
		ZReturnFalse(zm)
	}

	routine := NewZRoutine(zm.story, routineAddr, zm.pc)
	// copy operands to locals
	for i, v := range operands {
		routine.locals[i] = v
	}

	zm.pc = routineAddr
	zm.stack.Push(routine)
}

func ZReturn(zm *ZMachine, retValue uint16) {
	zm.pc = zm.stack.Top().retAddr
	zm.stack.Pop()

	zm.StoreReturn(retValue)
}

func ZReturnFalse(zm *ZMachine) {
	ZReturn(zm, uint16(0))
}

func ZReturnTrue(zm *ZMachine) {
	ZReturn(zm, uint16(1))
}
