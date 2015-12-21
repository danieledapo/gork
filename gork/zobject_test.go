package gork

import "testing"

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
	[]byte{0x04, 0x7E, 0x97, 0xC0, 0xA5, 0xB2, 0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4, 0x10, 0x82, 0x00},
	[]byte{0x04, 0x23, 0xC8, 0xC6, 0x95, 0xB2, 0x46, 0xDC, 0x42, 0xC2, 0x42, 0xB4, 0x00},
}

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
			16: []byte{0x82},
		},
	},
	ZObject{
		number:        3,
		attributes:    genAttrs(7, 22, 23),
		parent:        1,
		sibling:       2,
		child:         0,
		name:          "cyclop",
		propertiesPos: 0x0073,
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
		// default properties are all 0s in this case
		ret[i] = 0x00
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

func TestZObject(t *testing.T) {
	mem := ZMemory(createZObjectBuf())
	header := &ZHeader{objTblPos: 0x00}

	count := ZObjectsCount(&mem, header)
	if count != uint8(len(zobjectExpected)) {
		t.Fail()
	}

	for i := uint8(0); i < count; i++ {
		obj := NewZObject(&mem, i+1, header)
		expected := zobjectExpected[i]

		if obj.number != expected.number ||
			obj.attributes != expected.attributes ||
			obj.parent != expected.parent ||
			obj.sibling != expected.sibling ||
			obj.child != expected.child ||
			obj.name != expected.name ||
			obj.propertiesPos != expected.propertiesPos ||
			obj.mem != &mem {
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

func TestZObjectId(t *testing.T) {
	header := &ZHeader{objTblPos: 0}
	for i := byte(0); i < 255; i++ {
		addr := ZObjectAddress(i+1, header)

		if ZObjectId(addr, header) != i+1 {
			t.Fail()
		}
	}
}
