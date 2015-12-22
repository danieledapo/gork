package gork

import (
	"encoding/binary"
	"testing"
)

// attributes, parent, sibling, child
// DO NOT STORE propPos it will be calculated!
var zobjectData [][]byte = [][]byte{
	[]byte{0x00, 0x02, 0x00, 0x08, 0, 0, 2},
	[]byte{0x01, 0x00, 0x03, 0x00, 1, 0, 0},
	[]byte{0x01, 0x00, 0x03, 0x00, 1, 2, 0},
}

// nameLength, optional name, props
var zobjectProps [][]byte = [][]byte{
	[]byte{0x00, 0xB2, 0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4, 0x10, 0x82, 0x00},
	[]byte{0x02, 0x7E, 0x97, 0xC0, 0xA5, 0xB2, 0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4, 0x30, 0x82, 0x21, 0x00},
	[]byte{0x02, 0x23, 0xC8, 0xC6, 0x95, 0xB2, 0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4, 0x00},
}

const defaultPropByte byte = 0xFF
const defaultPropWord uint16 = uint16(defaultPropByte)<<8 | uint16(defaultPropByte)

// TODO generate automatically propertiesPos
var zobjectExpected []ZObject = []ZObject{
	ZObject{
		number:        1,
		attributes:    genAttrs(14, 28),
		parent:        0,
		sibling:       0,
		child:         2,
		name:          "",
		propertiesPos: 0x0059,
		properties: map[byte][]byte{
			18: []byte{0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4},
			16: []byte{0x82},
		},
	},
	ZObject{
		number:        2,
		attributes:    genAttrs(7, 22, 23),
		parent:        1,
		sibling:       0,
		child:         0,
		name:          "zork",
		propertiesPos: 0x0064,
		properties: map[byte][]byte{
			18: []byte{0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4},
			16: []byte{0x82, 0x21},
		},
	},
	ZObject{
		number:        3,
		attributes:    genAttrs(7, 22, 23),
		parent:        1,
		sibling:       2,
		child:         0,
		name:          "cyclop",
		propertiesPos: 0x0074,
		properties: map[byte][]byte{
			18: []byte{0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4},
		},
	},
}

func genAttrs(attrs ...int) [32]bool {
	var ret [32]bool

	for _, a := range attrs {
		ret[a] = true
	}

	return ret
}

func createZObjectBuf() []byte {
	// v3
	// 31 2 bytes properties
	ret := make([]byte, 31*2)

	for i := range ret {
		// default properties are all 0xFFs in this case
		ret[i] = defaultPropByte
	}

	firstPropPos := uint16(len(ret)) + uint16(len(zobjectData))*uint16(zobjectSize)

	lastPropPos := firstPropPos
	for i := range zobjectData {
		ret = append(ret, zobjectData[i]...)
		ret = append(ret, byte(lastPropPos>>8), byte(lastPropPos))

		lastPropPos += uint16(len(zobjectProps[i]))

	}

	for _, prop := range zobjectProps {
		ret = append(ret, prop...)
	}

	return ret
}

func prelude() (*ZMemory, *ZHeader, uint8) {
	mem := ZMemory(createZObjectBuf())
	header := &ZHeader{objTblPos: 0x00}

	count := ZObjectsCount(&mem, header)

	return &mem, header, count
}

func TestZObjectCount(t *testing.T) {
	_, _, count := prelude()

	if count != uint8(len(zobjectExpected)) {
		t.Fail()
	}
}

func TestZObject(t *testing.T) {
	mem, header, count := prelude()

	for i := uint8(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)
		expected := zobjectExpected[i]

		if obj.number != expected.number ||
			obj.attributes != expected.attributes ||
			obj.parent != expected.parent ||
			obj.sibling != expected.sibling ||
			obj.child != expected.child ||
			obj.name != expected.name ||
			obj.propertiesPos != expected.propertiesPos ||
			obj.mem != mem {
			t.Fail()
		}

		if len(obj.properties) != len(expected.properties) {
			t.Fail()
		}

		for k, v := range expected.properties {
			p, ok := obj.properties[k]
			if !ok {
				t.Fail()
			}

			if len(v) != len(p) {
				t.Fail()
			}

			for i := range p {
				if p[i] != v[i] {
					t.Fail()
				}
			}
		}

	}
}

func TestZObjectGetProperty(t *testing.T) {
	mem, header, count := prelude()

	for i := byte(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)
		good := zobjectExpected[i]

		for key, prop := range good.properties {
			if len(prop) <= 2 {
				rProp := obj.GetProperty(key)

				if len(prop) == 1 {
					if rProp != uint16(prop[0]) {
						t.Fail()
					}
				} else if rProp != binary.BigEndian.Uint16(prop) {
					t.Fail()
				}
			}
		}

		// should return default property
		if _, ok := obj.properties[31]; !ok && obj.GetProperty(31) != defaultPropWord {
			t.Fail()
		}
	}
}

func TestZObjectSetProperty(t *testing.T) {
	mem, header, count := prelude()

	for i := byte(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)

		var expected uint16

		for id, prop := range obj.properties {
			ok := true
			switch len(prop) {
			case 1:
				expected = uint16(defaultPropByte)
			case 2:
				expected = defaultPropWord
			default:
				ok = false
			}

			if ok {
				obj.SetProperty(id, expected)
				if obj.GetProperty(id) != expected {
					t.Fail()
				}
			}
		}
	}
}

func TestZObjectPropertyLen(t *testing.T) {
	mem, header, count := prelude()

	for i := byte(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)

		seq := mem.GetSequential(uint32(obj.propertiesPos))
		if seq.ReadByte() != 0 {
			// skip name
			seq.DecodeZString(header)
		}
		// skip dataSize
		seq.ReadByte()

		propertyPos := uint16(seq.pos)

		for _, k := range obj.PropertiesIds() {
			prop := obj.properties[k]

			if GetPropertyLen(mem, uint32(propertyPos)) != uint16(len(prop)) {
				t.Fail()
			}

			// skip propData and next dataSize
			propertyPos += uint16(len(prop)) + 1
		}
	}
}

func TestZObjectGetFirstPropertyAddr(t *testing.T) {
	mem, header, count := prelude()

	for i := byte(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)

		seq := mem.GetSequential(uint32(obj.propertiesPos))

		if seq.ReadByte() != 0 {
			// skip name
			seq.DecodeZString(header)
		}

		if obj.GetFirstPropertyAddr() != seq.pos {
			t.Fail()
		}
	}
}

func TestZObjectGetPropertyAddr(t *testing.T) {
	mem, header, count := prelude()

	for i := byte(0); i < count; i++ {
		obj := NewZObject(mem, i+1, header)

		seq := mem.GetSequential(uint32(obj.propertiesPos))

		if seq.ReadByte() != 0 {
			// skip name
			seq.DecodeZString(header)
		}
		propPos := seq.pos

		keys := obj.PropertiesIds()

		if len(keys) > 0 && obj.GetPropertyAddr(keys[0]) != obj.GetFirstPropertyAddr() {
			t.Fail()
		}

		// v3
		keyIdx := 0
		for propId := byte(31); propId > 0; propId-- {
			expected := uint32(0)
			if keyIdx < len(keys) && keys[keyIdx] == propId {
				expected = propPos
				keyIdx++

				// skip size
				propPos++
				propPos += uint32(len(zobjectExpected[i].properties[propId]))
			}

			if obj.GetPropertyAddr(byte(propId)) != expected {
				t.Fail()
			}

		}
	}
}

func TestZObjectId(t *testing.T) {
	header := &ZHeader{objTblPos: 0}
	for i := byte(0); i < 255; i++ {
		addr := ZObjectAddress(i+1, header)

		if ZObjectId(addr, header) != i+1 {
			t.Fail()
		}
	}
}
