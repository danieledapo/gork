package gork

import (
	"fmt"
)

const (
	MaxZObjects    = 255       // v3
	zobjectSize    = uint16(9) // v3
	propertyOffset = uint16(7) //v3
)

type ZObject struct {
	number        uint8
	attributes    []byte //v3 never gonna be more than 32
	parent        byte
	sibling       byte
	child         byte
	name          string
	propertiesPos uint16
	properties    map[byte][]byte
}

func NewZObject(story []byte, number uint8, objTblPos uint16, abbrTblPos uint16) *ZObject {
	obj := new(ZObject)
	obj.configure(story, number, objTblPos, abbrTblPos)
	return obj
}

func (obj *ZObject) configure(story []byte, number uint8, objTblPos uint16, abbrTblPos uint16) {
	obj.number = number

	addr := ZObjectAddress(number, objTblPos)

	// v3 attributes is 32 bit
	// more significant bit <-> attribute # smaller
	//
	// Bit  #  0 1 2 3 4 5 6 7
	// Attr #  7 6 5 4 3 2 1 0
	attr := ReadUint32(story, addr)
	for i := 3; i >= 0; i-- {
		byteno := uint8(3 - i)
		bits := uint8(attr >> (uint8(i) * 8))
		for j := 7; j >= 0; j-- {
			if (bits>>uint8(j))&0x01 == 0x01 {
				obj.attributes = append(obj.attributes, byteno*8+7-uint8(j))
			}
		}
	}

	obj.parent = ReadZByte(story, addr+4)
	obj.sibling = ReadZByte(story, addr+5)
	obj.child = ReadZByte(story, addr+6)
	obj.propertiesPos = ReadZWord(story, addr+propertyOffset)

	obj.readProperties(story, abbrTblPos)
}

func (obj *ZObject) readProperties(story []byte, abbrTblPos uint16) {
	// v3

	obj.properties = make(map[byte][]byte)

	// number of words
	textLength := uint16(ReadZByte(story, obj.propertiesPos))
	if textLength != 0 {
		obj.name = string(DecodeZString(story, obj.propertiesPos+1, abbrTblPos))
	}

	propPos := obj.propertiesPos + 1 + textLength*2

	dataSize := ReadZByte(story, propPos)
	for dataSize > 0 {
		prop := dataSize & (0x20 - 1)
		count := ((dataSize & 0xE0) >> 5) + 1
		propPos++

		for i := byte(0); i < count; i++ {
			obj.properties[prop] = append(obj.properties[prop],
				ReadZByte(story, propPos))
			propPos++
		}
		dataSize = ReadZByte(story, propPos)
	}
}

func ZObjectAddress(idx uint8, objTblPos uint16) uint16 {
	if idx < 1 {
		panic("objects are numbered from 1 to 255")
	}
	// v3 skip 31 words containing property default table
	return uint16(objTblPos) + 31*2 + uint16(idx-1)*uint16(zobjectSize)
}

func ZObjectsCount(story []byte, objTblPos uint16) uint8 {
	count := uint8(0)
	firstPropertyPos := uint16(0)
	addr := uint16(0)

	doCount := func() {
		count++
		addr = ZObjectAddress(count, objTblPos)

		if firstPropertyPos == 0 || addr < firstPropertyPos {
			addr += propertyOffset

			propertyPos := ReadZWord(story, addr)
			if firstPropertyPos == 0 || propertyPos < firstPropertyPos {
				firstPropertyPos = propertyPos
			}
		}
	}

	// v3 objects tree
	// object #1
	// object #2
	// ...
	// object #N
	// object #1 properties
	doCount()
	for addr < firstPropertyPos {
		doCount()
	}

	// do not count object #0
	return count - 1
}

func (obj *ZObject) String() string {
	ret := ""

	ret += fmt.Sprintf("Attributes: ")
	if len(obj.attributes) > 0 {
		// ret += fmt.Sprintf("%d\n", obj.attributes)
		for _, attr := range obj.attributes {
			ret += fmt.Sprintf("%d, ", attr)
		}
		// do not include " ,"
		ret = ret[:len(ret)-2]
		ret += fmt.Sprintln("")
	} else {
		ret += fmt.Sprintln("None\n")
	}

	ret += fmt.Sprintf("     Parent object: %3d  ", obj.parent)
	ret += fmt.Sprintf("Sibling object: %3d  ", obj.sibling)
	ret += fmt.Sprintf("Child object: %3d\n", obj.child)

	ret += fmt.Sprintf("     Property address: %04x\n", obj.propertiesPos)
	ret += fmt.Sprintf("         Description: \"%s\"\n", obj.name)

	ret += fmt.Sprintln("          Properties:")
	for k, v := range obj.properties {
		ret += fmt.Sprintf("              [%2d] ", k)
		for b := range v {
			ret += fmt.Sprintf("%02X ", v[b])
		}
		ret += fmt.Sprintln("")
	}
	ret += fmt.Sprintln("")

	return ret
}

func DumpAllZObjects(story []byte, objTblPos uint16, abbrTblPos uint16) {
	total := ZObjectsCount(story, objTblPos)

	fmt.Print("\n    **** Objects ****\n\n")
	fmt.Printf("  Object count = %d\n", total)

	for i := uint8(1); i <= total; i++ {
		fmt.Printf("%3d. %s", i, NewZObject(story, i, objTblPos, abbrTblPos))
	}
}

func DumpZObjectsTree(story []byte, objTblPos uint16, abbrTblPos uint16) {

	fmt.Print("\n    **** Object tree ****\n\n")

	total := ZObjectsCount(story, objTblPos)

	var printObject func(obj *ZObject, depth int)
	printObject = func(obj *ZObject, depth int) {
		for {

			for j := 0; j < depth; j++ {
				fmt.Print(" . ")
			}
			fmt.Printf("[%3d] ", obj.number)
			fmt.Printf("\"%s\"\n", obj.name)

			if obj.child != 0 {
				childobj := NewZObject(story, obj.child, objTblPos, abbrTblPos)
				printObject(childobj, depth+1)
			}

			if obj.sibling == 0 {
				break
			}
			obj = NewZObject(story, obj.sibling, objTblPos, abbrTblPos)
		}
	}

	for i := uint8(1); i <= total; i++ {
		zobj := NewZObject(story, i, objTblPos, abbrTblPos)

		// root
		if zobj.parent == 0 {
			printObject(zobj, 0)
		}
	}
}
