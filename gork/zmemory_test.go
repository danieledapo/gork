package gork

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

var header *ZHeader = &ZHeader{
	abbrTblPos: 2,
}

// len must be multiple of 4
var readTestData []byte = []byte{42, 73, 96, 7, 28, 1, 2, 3}
var writeTestData []byte = []byte{3, 2, 1, 28, 7, 96, 73, 42}
var zstrings [][]byte = [][]byte{
	[]byte{0x7E, 0x97, 0xC0, 0xA5},
	[]byte{0x23, 0xC8, 0xC6, 0x95},
	[]byte{0x80, 0xA5},
	[]byte{0x84, 0x05, 0x00, 0x02, 0x7E, 0x97, 0xC0, 0xA5},
}
var zstringsExpected []string = []string{
	"zork",
	"cyclop",
	" ",
	"zork",
}

var encodedZstrings []string = []string{
	"zork",
	"cyclop",
	"i",
	"42,",
	"$",
}
var encodedZstringsExpected [][]uint16 = [][]uint16{
	[]uint16{0x7E97, 0xC0A5},
	[]uint16{0x23C8, 0xC695},
	[]uint16{0x38A5, 0x94A5},
	[]uint16{0x1585, 0xA8B3},
	[]uint16{0x16A5, 0x94A5},
}

var byteOrder binary.ByteOrder = binary.BigEndian

func TestByteAt(t *testing.T) {
	mem := ZMemory(readTestData)

	for i := range readTestData {
		if readTestData[i] != mem.ByteAt(uint32(i)) {
			t.Fail()
		}
	}
}

func TestWordAt(t *testing.T) {
	mem := ZMemory(readTestData)

	for i := uint32(0); i < uint32(len(readTestData)/2); i++ {
		if byteOrder.Uint16(mem[i:i+2]) != mem.WordAt(i) {
			t.Fail()
		}
	}
}

func TestUint32At(t *testing.T) {
	mem := ZMemory(readTestData)

	for i := uint32(0); i < uint32(len(readTestData)/4); i++ {
		if byteOrder.Uint16(mem[i:i+4]) != mem.WordAt(i) {
			t.Fail()
		}
	}
}

func TestWriteByteAt(t *testing.T) {
	mem := ZMemory(readTestData)

	for i := range readTestData {
		mem.WriteByteAt(uint32(i), writeTestData[i])
		if mem.ByteAt(uint32(i)) != writeTestData[i] {
			t.Fail()
		}
	}
}

func TestWriteWordAt(t *testing.T) {
	mem := ZMemory(readTestData)

	for i := uint32(0); i < uint32(len(readTestData)/2); i++ {
		toWrite := byteOrder.Uint16(writeTestData[i : i+2])
		mem.WriteWordAt(i, toWrite)
		if toWrite != mem.WordAt(i) {
			t.Fail()
		}
	}
}

func TestPeekByte(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := range readTestData {
		if seq.PeekByte() != seq.mem.ByteAt(seq.pos) || seq.pos != uint32(i) {
			t.Fail()
		}
		seq.pos += 1
	}
}

func TestPeekWord(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := uint32(0); i < uint32(len(mem)/2); i++ {
		if seq.PeekWord() != seq.mem.WordAt(seq.pos) || seq.pos != uint32(i*2) {
			t.Fail()
		}
		seq.pos += 2
	}
}

func TestPeekUint32(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := uint32(0); i < uint32(len(mem)/4); i++ {
		if seq.PeekUInt32() != seq.mem.UInt32At(seq.pos) || seq.pos != uint32(i*4) {
			t.Fail()
		}
		seq.pos += 4
	}
}

func TestReadByte(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := range readTestData {
		if seq.pos != uint32(i) || seq.ReadByte() != seq.mem.ByteAt(uint32(i)) {
			t.Fail()
		}
	}
}

func TestReadWord(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := uint32(0); i < uint32(len(mem)/2); i++ {
		if seq.pos != uint32(i*2) || seq.ReadWord() != seq.mem.WordAt(i*2) {
			t.Fail()
		}
	}
}

func TestReadUint32(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := uint32(0); i < uint32(len(mem)/4); i++ {
		if seq.pos != uint32(i*4) || seq.ReadUint32() != seq.mem.UInt32At(i*4) {
			t.Fail()
		}
	}
}

func TestWriteByte(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := range readTestData {
		seq.WriteByte(writeTestData[i])
		if mem.ByteAt(uint32(i)) != writeTestData[i] || seq.pos != uint32(i+1) {
			t.Fail()
		}
	}
}

func TestWriteWord(t *testing.T) {
	mem := ZMemory(readTestData)
	seq := mem.GetSequential(0)

	for i := uint32(0); i < uint32(len(readTestData)/2); i++ {
		if seq.pos != i*2 {
			t.Fail()
		}

		toWrite := byteOrder.Uint16(writeTestData[i : i+2])
		seq.WriteWord(toWrite)

		if toWrite != mem.WordAt(i*2) {
			fmt.Println("hey")

			t.Fail()
		}
	}
}

func TestPackedAddres(t *testing.T) {
	for i := uint32(0); i < 10; i++ {
		if !IsPackedAddress(PackedAddress(i)) {
			t.Fail()
		}
	}
}

func TestZStringDecodeAt(t *testing.T) {
	for i, zstring := range zstrings {
		mem := ZMemory(zstring)

		// in this case zstring doesn't have abbreviations,
		// so don't pass the header
		if mem.DecodeZStringAt(0, header) != zstringsExpected[i] {
			t.Fail()
		}
	}
}

func TestZStringDecode(t *testing.T) {
	for _, zstring := range zstrings {
		mem := ZMemory(zstring)
		seq := mem.GetSequential(0)

		if mem.DecodeZStringAt(0, header) != seq.DecodeZString(header) {
			// cannot be sure where seq.pos will be

			t.Fail()
		}
	}
}

func TestZStringEncode(t *testing.T) {
	for i, zstr := range encodedZstrings {
		expected := encodedZstringsExpected[i]
		encoded := ZStringEncode(zstr)

		for i := range encoded {
			if encoded[i] != expected[i] {
				fmt.Printf("%X %X\n", encoded[i], expected[i])
				t.Fail()
			}

			buf := make([]byte, len(encoded)*2)
			for i, v := range encoded {
				buf[i*2] = byte(v >> 8)
				buf[i*2+1] = byte(v)
			}

			seq := ZMemory(buf)
			decoded := seq.DecodeZStringAt(0, nil)

			for j := range decoded {
				ch := decoded[j]

				// default value if not found
				expCh := byte('?')

				for _, alph := range Alphabets {
					if strings.IndexByte(alph, ch) >= 0 {
						expCh = ch
						break
					}
				}

				if ch != expCh {
					t.Fail()
				}

			}
		}

	}
}
