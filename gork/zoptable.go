package gork

import (
	"fmt"
	"log"
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
	nil,
	nil,
	nil,
	nil,
	ZRetPop,
	nil,
	nil,
	ZNl,
}

var oneOpFuncs = []OneOpFunc{
	ZJ0,
	ZGetSibling,
	ZGetChild,
	ZGetParent,
	nil,
	ZInc,
	nil,
	ZPrintAt,
	nil,
	nil,
	ZPrintObject,
	ZReturn,
	ZJump,
	ZPrintAtPacked,
	ZLoad,
	ZNot,
}

var twoOpFuncs = []TwoOpFunc{
	ZNOOP,
	nil, // ZJe is a two op func but it accepts VAR count of args
	ZJl,
	ZJg,
	nil,
	ZIncChk,
	ZJin,
	nil,
	ZOr,
	ZAnd,
	ZTestAttr,
	ZSetAttr,
	nil,
	ZStore,
	ZInsertObj,
	ZLoadW,
	ZLoadB,
	ZGetProp,
	nil,
	nil,
	ZAdd,
	ZSub,
	ZMul,
	ZDiv,
	ZMod,
}

var varOpFuncs = []VarOpFunc{
	ZCall,
	ZStoreW,
	ZStoreB,
	ZPutProp,
	nil,
	ZPrintChar,
	ZPrintNum,
	nil,
	ZPush,
	ZPull,
}

func ZCall(zm *ZMachine, operands []uint16) {
	routineAddr := PackedAddress(uint32(operands[0]))

	retAddr := zm.seq.pos
	zm.seq.pos = routineAddr
	routine := NewZRoutine(zm.seq, retAddr)

	zm.stack.Push(routine)

	if routineAddr == 0 {
		ZReturnFalse(zm)
		return
	}

	if len(operands) > 1 {
		// copy operands to locals
		for i, v := range operands[1:] {
			routine.locals[i] = v
		}
	}
	log.Print("Call ", routine)
}

func ZReturn(zm *ZMachine, retValue uint16) {
	zm.seq.pos = zm.stack.Pop().retAddr
	log.Printf("Returning to 0x%X\n", zm.seq.pos)
	zm.StoreReturn(retValue)
}

func ZReturnFalse(zm *ZMachine) {
	ZReturn(zm, uint16(0))
}

func ZReturnTrue(zm *ZMachine) {
	ZReturn(zm, uint16(1))
}

func ZJe(zm *ZMachine, args []uint16) {
	conditionOk := false
	for _, v := range args[1:] {
		if v == args[0] {
			conditionOk = true
			break
		}
	}
	zm.Branch(conditionOk)
}

func ZJl(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.Branch(lhs < rhs)
}

func ZJg(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.Branch(lhs > rhs)
}

func ZJ0(zm *ZMachine, op uint16) {
	zm.Branch(op == 0)
}

func ZJump(zm *ZMachine, offset uint16) {
	// uncoditional branch
	// this is not a branch instruction
	// jumping to an instruction in a different routine is permitted,
	// but the standard consider it bad practice :)
	zm.seq.pos = zm.CalcJumpAddress(int16(offset))
}

func ZPrint(zm *ZMachine) {
	str := zm.seq.DecodeZString(zm.header)
	fmt.Print(str)
}

func ZPrintRet(zm *ZMachine) {
	ZPrint(zm)
	fmt.Println("")
	ZReturnTrue(zm)
}

func ZPrintObject(zm *ZMachine, obj uint16) {
	// objects are 1-based
	fmt.Print(zm.objects[obj-1].name)
}

func ZPrintAt(zm *ZMachine, addr uint16) {
	str := zm.seq.mem.DecodeZStringAt(uint32(addr), zm.header)
	fmt.Print(str)
}

func ZPrintAtPacked(zm *ZMachine, paddr uint16) {
	str := zm.seq.mem.DecodeZStringAt(PackedAddress(uint32(paddr)), zm.header)
	fmt.Print(str)
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

func ZAdd(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.StoreReturn(lhs + rhs)
}

func ZSub(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.StoreReturn(lhs - rhs)
}

func ZMul(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.StoreReturn(lhs * rhs)
}

func ZDiv(zm *ZMachine, lhs uint16, rhs uint16) {
	if rhs == 0 {
		log.Fatal("division by zero error")
	}
	zm.StoreReturn(lhs / rhs)
}

func ZMod(zm *ZMachine, lhs uint16, rhs uint16) {
	if rhs == 0 {
		log.Fatal("mod by zero error")
	}
	zm.StoreReturn(lhs % rhs)
}

func ZOr(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.StoreReturn(lhs | rhs)
}

func ZAnd(zm *ZMachine, lhs uint16, rhs uint16) {
	zm.StoreReturn(lhs & rhs)
}

// v3
func ZNot(zm *ZMachine, arg uint16) {
	zm.StoreReturn(^arg)
}

func ZNOOP(_ *ZMachine, _ uint16, _ uint16) {
	log.Fatal("NO OP 2OP")
}

func ZLoad(zm *ZMachine, varnum uint16) {
	zm.StoreReturn(zm.GetVarAt(byte(varnum)))
}

func ZLoadB(zm *ZMachine, array uint16, bidx uint16) {
	// TODO access violation
	zm.StoreReturn(uint16(zm.seq.mem.ByteAt(uint32(array + bidx))))
}

func ZLoadW(zm *ZMachine, array uint16, widx uint16) {
	// TODO access violation
	// index is the index of the nth word
	zm.StoreReturn(zm.seq.mem.WordAt(uint32(array + widx*2)))
}

func ZStore(zm *ZMachine, varnum uint16, value uint16) {
	zm.StoreVarAt(byte(varnum), value)
}

func ZStoreB(zm *ZMachine, args []uint16) {
	// TODO access violation
	addr := args[0] + args[1]
	zm.seq.mem.WriteByteAt(uint32(addr), byte(args[2]))
}

func ZStoreW(zm *ZMachine, args []uint16) {
	// TODO access violation
	// index is the index of the nth word
	addr := uint32(args[0]) + uint32(args[1])*2
	zm.seq.mem.WriteWordAt(addr, args[2])
}

func ZPush(zm *ZMachine, args []uint16) {
	zm.stack.Top().locals = append(zm.stack.Top().locals, args[0])
}

func ZPull(zm *ZMachine, args []uint16) {
	varnum := args[0]
	zm.stack.Top().locals = append(zm.stack.Top().locals[:varnum],
		zm.stack.Top().locals[varnum+1:]...)
	// should not zm.StoreReturn popped value
}

func ZPop(zm *ZMachine) {
	zm.stack.Top().locals = zm.stack.Top().locals[:len(zm.stack.Top().locals)-1]
}

func ZRetPop(zm *ZMachine) {
	ret := zm.stack.Top().locals[len(zm.stack.Top().locals)-1]
	ZPop(zm)
	ZReturn(zm, ret)
}

func ZInsertObj(zm *ZMachine, objectId uint16, newParentId uint16) {
	zm.ResetObjectParent(uint8(objectId), uint8(newParentId))
}

func ZJin(zm *ZMachine, childId uint16, parentId uint16) {
	condition := zm.objects[childId-1].parent == uint8(parentId)
	zm.Branch(condition)
}

func ZGetSibling(zm *ZMachine, objectId uint16) {
	sibling := zm.objects[objectId-1].sibling
	zm.StoreReturn(uint16(sibling))
	zm.Branch(sibling != NULL_OBJECT_INDEX)
}

func ZGetChild(zm *ZMachine, objectId uint16) {
	child := zm.objects[objectId-1].child
	zm.StoreReturn(uint16(child))
	zm.Branch(child != NULL_OBJECT_INDEX)
}

func ZGetParent(zm *ZMachine, objectId uint16) {
	zm.StoreReturn(uint16(zm.objects[objectId-1].parent))
}

func ZPutProp(zm *ZMachine, args []uint16) {
	zm.objects[args[0]-1].SetProperty(byte(args[1]), args[2])
}

func ZGetProp(zm *ZMachine, objectId uint16, propertyId uint16) {
	res := zm.objects[objectId-1].GetProperty(byte(propertyId))
	zm.StoreReturn(res)
}

func ZGetPropLen(zm *ZMachine, propertyAddr uint16) {
	if propertyAddr == 0 {
		zm.StoreReturn(0)
	} else {
		res := GetPropertyLen(zm.seq.mem, uint32(propertyAddr))
		zm.StoreReturn(res)
	}
}

func ZGetPropAddr(zm *ZMachine, objectId uint16, propertyId uint16) {
	addr := zm.objects[objectId-1].GetPropertyAddr(byte(propertyId))
	zm.StoreReturn(uint16(addr))
}

func ZTestAttr(zm *ZMachine, objectId uint16, attrId uint16) {
	cond := zm.objects[objectId-1].attributes[attrId]
	zm.Branch(cond)
}

func ZSetAttr(zm *ZMachine, objectId uint16, attrId uint16) {
	zm.objects[objectId-1].attributes[attrId] = true
}

func ZClearAttr(zm *ZMachine, objectId uint16, attrId uint16) {
	zm.objects[objectId-1].attributes[attrId] = false
}

func ZNl(_ *ZMachine) {
	fmt.Println("")
}

func ZInc(zm *ZMachine, varnum uint16) {
	zm.UpdateVarAt(byte(varnum), +1)
}

func ZDec(zm *ZMachine, varnum uint16) {
	zm.UpdateVarAt(byte(varnum), -1)
}

func ZIncChk(zm *ZMachine, varnum uint16, value uint16) {
	newValue := zm.UpdateVarAt(byte(varnum), +1)
	zm.Branch(int16(newValue) > int16(value))
}

func ZDecChk(zm *ZMachine, varnum uint16, value uint16) {
	newValue := zm.UpdateVarAt(byte(varnum), -1)
	zm.Branch(int16(newValue) < int16(value))
}
