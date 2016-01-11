package gork

import (
	"fmt"
	"log"
	"sort"
)

const (
	MaxZObjects       = 255       // v3
	zobjectSize       = uint32(9) // v3
	propertyOffset    = uint32(7) //v3
	NULL_OBJECT_INDEX = uint8(0)
)

type ZObject struct {
	number        uint8
	attributes    [32]bool //v3 never gonna be more than 32
	parent        byte
	sibling       byte
	child         byte
	name          string
	propertiesPos uint16
	properties    map[byte][]byte
	mem           *ZMemory
	header        *ZHeader
}

func NewZObject(mem *ZMemory, number uint8, header *ZHeader) *ZObject {
	obj := new(ZObject)
	obj.configure(mem, number, header)
	return obj
}

func (obj *ZObject) configure(mem *ZMemory, number uint8, header *ZHeader) {
	obj.number = number

	obj.mem = mem
	obj.header = header

	seq := mem.GetSequential(ZObjectAddress(number, header))

	// v3 attributes is 32 bit
	// more significant bit <-> attribute # smaller
	//
	// Bit  #  0 1 2 3 4 5 6 7
	// Attr #  7 6 5 4 3 2 1 0
	attr := seq.ReadUint32()
	for i := 3; i >= 0; i-- {
		byteno := uint8(3 - i)
		bits := uint8(attr >> (uint8(i) * 8))
		for j := 7; j >= 0; j-- {
			obj.attributes[byteno*8+7-uint8(j)] = (bits>>uint8(j))&0x01 == 0x01
		}
	}

	obj.parent = seq.ReadByte()
	obj.sibling = seq.ReadByte()
	obj.child = seq.ReadByte()
	obj.propertiesPos = seq.ReadWord()

	obj.readProperties(header)
}

func (obj *ZObject) readProperties(header *ZHeader) {
	// v3

	obj.properties = make(map[byte][]byte)

	seq := obj.mem.GetSequential(uint32(obj.propertiesPos))

	// number of words
	textLength := uint16(seq.ReadByte())
	if textLength != 0 {
		obj.name = string(seq.DecodeZString(header))
	}

	dataSize := seq.ReadByte()
	for dataSize > 0 {
		prop := dataSize & (0x20 - 1)
		count := ((dataSize & 0xE0) >> 5) + 1

		for i := byte(0); i < count; i++ {
			obj.properties[prop] = append(obj.properties[prop],
				seq.ReadByte())
		}
		dataSize = seq.ReadByte()
	}
}

func (obj *ZObject) SetProperty(propertyId byte, value uint16) {
	if _, ok := obj.properties[propertyId]; !ok {
		log.Fatalf("Property %d not found\n", propertyId)
	}

	// TODO
	// check if it is legal for a story to get/set property
	// directly via store/load or it is forced to use putprop/getprop
	// in case it's legal try to uncomment the commented lines and hope
	// the best
	// however if the story table address can be modified as well
	// than ZMachine cannot cache objects, because otherwise it would
	// be too complex to intercept writes to the address of the property
	// table of an object

	// addr := obj.GetPropertyAddr(propertyId)

	// it seems we cannot take address of values of a map :'(
	switch len(obj.properties[propertyId]) {
	case 1:
		// store only least significant byte
		data := byte(value & 0x00FF)
		obj.properties[propertyId][0] = data
		// obj.mem.WriteByteAt(addr, data)
	case 2:
		// big endian
		obj.properties[propertyId][0] = byte((value >> 8) & 0x00FF)
		obj.properties[propertyId][1] = byte(value & 0x00FF)
		// obj.mem.WriteWordAt(addr, value)
	default:
		log.Fatal("cannot set property, because its length is > 2 bytes")
	}
}

func (obj *ZObject) GetProperty(propertyId byte) uint16 {
	if _, ok := obj.properties[propertyId]; !ok {
		// DON'T PANIC, cause the property could be in the
		// global default properties table

		// v3
		if propertyId < 1 || propertyId > 31 {
			log.Fatalf("Invalid propertyIndex %d, values range in v3 is [1,31]\n", propertyId)
		}

		// property table is a sequence of words
		addr := uint32(obj.header.objTblPos) + uint32((propertyId-1)*2)
		return obj.mem.WordAt(addr)
	}

	res := uint16(0)

	// it seems we cannot take address of values of a map :'(
	switch len(obj.properties[propertyId]) {
	case 1:
		res = uint16(obj.properties[propertyId][0])
	case 2:
		// big endian
		res |= uint16(obj.properties[propertyId][0]) << 8
		res |= uint16(obj.properties[propertyId][1])
	default:
		log.Fatal("cannot get property, because its length is > 2 bytes")
	}

	return res
}

func GetPropertyLen(mem *ZMemory, propertyPos uint32) uint16 {
	// the property size byte is the byte before propertyPos
	size := mem.ByteAt(uint32(propertyPos - 1))
	nbytes := (size >> 5) + 1
	return uint16(nbytes)
}

func (obj *ZObject) GetFirstPropertySizeAddr() uint32 {
	// returns the address of the size byte

	// text length is in words
	textLength := obj.mem.ByteAt(uint32(obj.propertiesPos))
	return uint32(obj.propertiesPos) + 1 + uint32(textLength)*2
}

func (obj *ZObject) GetPropertyAddr(propertyId byte) uint32 {
	// v3
	addr := obj.GetFirstPropertySizeAddr()

	for {
		size := obj.mem.ByteAt(addr)

		propno := size & 0x1F

		if size == 0 || propno < propertyId {
			// must return 0 if property is not present
			// properties are sorted in descending order
			return 0
		}

		// skip size
		addr++
		if propno == propertyId {
			return addr
		}

		addr += uint32((size >> 5) + 1)
	}
}

func (obj *ZObject) MakeOrphan(other []*ZObject) {
	// TODO
	// the same doubts of ZObject get/set properties apply here
	// refer to the comment in zobject.go:SetProperty for
	// better understanding

	if obj.parent != NULL_OBJECT_INDEX {
		parent := other[obj.parent-1]
		if parent.child == obj.number {
			// obj is the first child so move to sibling
			parent.child = obj.sibling
		} else {
			// we are among the siblings so update previous one
			curChildId := parent.child
			prevChildId := NULL_OBJECT_INDEX

			for curChildId != obj.number && curChildId != NULL_OBJECT_INDEX {
				prevChildId = curChildId
				curChildId = other[curChildId-1].sibling
			}
			// TODO
			// sanity checks

			// update sibling to next one
			other[prevChildId-1].sibling = obj.sibling
		}
	}
	obj.parent = NULL_OBJECT_INDEX
}

func (obj *ZObject) ChangeParent(newParentId uint8, other []*ZObject) {
	if obj.number == newParentId {
		log.Fatal("trying to set object's parent to the object itself,",
			"not sure is allowed")
	}

	obj.MakeOrphan(other)

	// change object so that its sibling is the first child of parent
	// set parent's child to objectId
	// set child's parent to the newParent
	other[obj.number-1].sibling = other[newParentId-1].child
	other[newParentId-1].child = obj.number
	other[obj.number-1].parent = newParentId
}

func (obj *ZObject) NextProperty(prop byte) byte {
	// props are sorted in descending order

	if prop == 0 {
		max := byte(0)
		for p, _ := range obj.properties {
			if p > max {
				max = p
			}
		}
		return max
	} else {
		nextMaxProp := byte(0)
		for p, _ := range obj.properties {
			if p < prop && p > nextMaxProp {
				nextMaxProp = p
			}
		}
		return nextMaxProp
	}
}

func ZObjectAddress(idx uint8, header *ZHeader) uint32 {
	if idx < 1 {
		log.Fatal("objects are numbered from 1 to 255")
	}
	// v3 skip 31 words containing property default table
	return uint32(header.objTblPos) + 31*2 + uint32(idx-1)*zobjectSize
}

func ZObjectId(address uint32, header *ZHeader) uint8 {
	res := (address - uint32(header.objTblPos) - 31*2) / zobjectSize
	return uint8(res) + 1
}

func ZObjectsCount(mem *ZMemory, header *ZHeader) uint8 {
	count := uint8(0)
	firstPropertyPos := uint32(0)

	seq := mem.GetSequential(ZObjectAddress(1, header))

	doCount := func() {
		count++
		seq.pos = ZObjectAddress(count, header)

		if firstPropertyPos == 0 || seq.pos < firstPropertyPos {
			seq.pos += propertyOffset

			propertyPos := uint32(seq.PeekWord())
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
	for seq.pos < firstPropertyPos {
		doCount()
	}

	// do not count object #0
	return count - 1
}

func (obj *ZObject) PropertiesIds() []byte {
	var keys []int
	for k := range obj.properties {
		keys = append(keys, int(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	ret := make([]byte, len(keys))

	for i, v := range keys {
		ret[i] = byte(v)
	}

	return ret
}

func (obj *ZObject) Id() uint8 {
	return obj.number
}

func (obj *ZObject) Name() string {
	return obj.name
}

func (obj *ZObject) ParentId() uint8 {
	return obj.parent
}

func (obj *ZObject) SiblingId() uint8 {
	return obj.sibling
}

func (obj *ZObject) ChildId() uint8 {
	return obj.child
}

func (obj *ZObject) String() string {
	ret := ""

	ret += fmt.Sprintf("Attributes: ")
	if len(obj.attributes) > 0 {
		for i, attr := range obj.attributes {
			if attr {
				ret += fmt.Sprintf("%d, ", i)
			}
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

	for _, k := range obj.PropertiesIds() {
		ret += fmt.Sprintf("              [%2d] ", k)
		for b := range obj.properties[k] {
			ret += fmt.Sprintf("%02X ", obj.properties[k][b])
		}
		ret += fmt.Sprintln("")
	}
	ret += fmt.Sprintln("")

	return ret
}
