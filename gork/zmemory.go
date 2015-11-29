package gork

import (
	"fmt"
)

type ZMemory []byte
type ZMemorySequential struct {
	mem *ZMemory
	pos uint16
}

func NewZMemory(mem []byte) *ZMemory {
	zmem := new(ZMemory)
	*zmem = mem
	return zmem
}

func (zmem *ZMemory) ByteAt(addr uint16) byte {
	return (*zmem)[addr]
}

func (zmem *ZMemory) WordAt(addr uint16) uint16 {
	// Big Endian
	return (uint16((*zmem)[addr]) << 8) |
		(uint16((*zmem)[addr+1]))
}

func (zmem *ZMemory) UInt32At(addr uint16) uint32 {
	// Big Endian
	return (uint32((*zmem)[addr]) << 24) |
		(uint32((*zmem)[addr+1]) << 16) |
		(uint32((*zmem)[addr+2]) << 8) |
		uint32((*zmem)[addr+3])
}

func (zmem *ZMemory) WriteByteAt(addr uint16, val byte) {
	(*zmem)[addr] = val
}

func (zmem *ZMemory) WriteWordAt(addr uint16, val uint16) {
	(*zmem)[addr] = byte(val & 0xF0)
	(*zmem)[addr+1] = byte(val & 0X0F)
}

func (zmem *ZMemory) GetSequential(addr uint16) *ZMemorySequential {
	return &ZMemorySequential{zmem, addr}
}

func (zmem *ZMemorySequential) PeekByte() byte {
	return zmem.mem.ByteAt(zmem.pos)
}

func (zmem *ZMemorySequential) PeekWord() uint16 {
	return zmem.mem.WordAt(zmem.pos)
}

func (zmem *ZMemorySequential) PeekUInt32() uint32 {
	return zmem.mem.UInt32At(zmem.pos)
}

func (zmem *ZMemorySequential) ReadByte() byte {
	tmp := zmem.mem.ByteAt(zmem.pos)
	zmem.pos++
	return tmp
}

func (zmem *ZMemorySequential) ReadWord() uint16 {
	tmp := zmem.mem.WordAt(zmem.pos)
	zmem.pos += 2
	return tmp
}

func (zmem *ZMemorySequential) ReadUint32() uint32 {
	tmp := zmem.mem.UInt32At(zmem.pos)
	zmem.pos += 4
	return tmp
}

func (zmem *ZMemory) DecodeZStringAt(addr uint16, header *ZHeader) string {
	return zmem.GetSequential(addr).DecodeZString(header)
}

func (zmem *ZMemorySequential) DecodeZString(header *ZHeader) string {
	// v3

	ret := ""
	data := uint16(0)
	code := uint16(0)

	alphabet := uint8(0)
	shiftLock := uint8(0)

	synonimFlag := false
	synonim := uint16(0)

	// 0 not ascii
	// 1 first part
	// 2 last part
	asciiPart := uint8(0)
	asciiFirstPart := uint16(0)

	for data&0x8000 == 0 {
		data = zmem.ReadWord()

		for i := 10; i >= 0; i -= 5 {
			code = (data >> uint8(i)) & 0x1F

			if synonimFlag {
				synonimFlag = false
				synonim = (synonim - 1) * 64

				tmpAddr := zmem.mem.WordAt(header.abbrTblPos+synonim+code*2) * 2
				ret += zmem.mem.DecodeZStringAt(tmpAddr, header)

				alphabet = shiftLock
			} else if asciiPart > 0 {
				tmp := asciiPart
				asciiPart++
				if tmp == 1 {
					asciiFirstPart = code << 5
				} else {
					asciiPart = 0
					ret += string(asciiFirstPart | code)
				}
			} else if code > 5 {
				code -= 6

				if alphabet == 2 && code == 0 {
					asciiPart = 1
				} else if alphabet == 2 && code == 1 {
					ret += "\n"
				} else {
					ret += string(Alphabets[alphabet][code])
				}
				alphabet = shiftLock
			} else if code == 0 {
				ret += " "
			} else if code < 4 {
				synonimFlag = true
				synonim = code
			} else {
				alphabet = uint8(code - 3)
				shiftLock = 0
			}
		}
	}

	return ret
}

func PackedAddress(addr uint16) uint16 {
	// v3
	return addr * 2
}

func IsPackedAddress(addr uint16) bool {
	// v3
	return addr%2 == 0
}

func (zmem *ZMemory) String() string {
	return fmt.Sprintf("buf: %s\n", zmem)
}