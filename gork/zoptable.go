package gork

import (
	"fmt"
)

type ZeroOpFunc func(*ZMachine)
type OneOpFunc func(*ZMachine, uint16)
type TwoOpFunc func(*ZMachine, uint16, uint16)
type VarOpFunc func(*ZMachine, []uint16)

var zeroOpFuncs = []ZeroOpFunc{
	ZReturnTrue,
	ZReturnFalse,
	ZPrint,
	ZPrintRet,
}

var oneOpFuncs = []OneOpFunc{
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	nil,
	ZPrintAt,
	nil,
	nil,
	ZPrintObject,
	ZReturn,
	nil,
	ZPrintAtPacked,
}

var twoOpFuncs = []TwoOpFunc{}

var varOpFuncs = []VarOpFunc{
	ZCall,
	nil,
	nil,
	nil,
	nil,
	ZPrintChar,
	ZPrintNum,
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

func ZPrint(zm *ZMachine) {
	str := DecodeZString(zm.story, zm.header)
	zm.pc += zm.story.pos
	fmt.Print(str)
}

func ZPrintRet(zm *ZMachine) {
	ZPrint(zm)
	fmt.Println("")
	ZReturnTrue(zm)
}

func ZPrintObject(zm *ZMachine, obj uint16) {
	fmt.Print(zm.objects[obj].name)
}

func ZPrintAt(zm *ZMachine, addr uint16) {
	str := DecodeZStringAt(zm.story, addr, zm.header)
	fmt.Print(str)
}

func ZPrintAtPacked(zm *ZMachine, paddr uint16) {
	ZPrintAt(zm, PackedAddress(paddr))
}

func ZPrintNum(zm *ZMachine, args []uint16) {
	fmt.Print(args[0])
}

func ZPrintChar(zm *ZMachine, args []uint16) {
	// print only ASCII
	if args[0] == 13 {
		fmt.Println("")
	} else if args[0] >= 32 && args[0] <= 126 {
		fmt.Printf("%c", args[0])
	} // ignore everything else
}
