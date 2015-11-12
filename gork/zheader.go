package gork

import (
	"fmt"
)

const (
	SerialSize = 6
)

// dynamic memory range: [0, dynMemSize)
// static memory range: [dynMemSize, min(fileLength, 0xFFFF)], actually the end is useless
// high memory range: [highStart, EOF)
type ZHeader struct {
	config       byte
	version      byte
	release      uint16
	highStart    uint16
	pc           uint16
	dictPos      uint16
	objTblPos    uint16
	globalsPos   uint16
	dynMemSize   uint16
	serial       [SerialSize]byte
	abbrTblPos   uint16
	fileLength   uint64
	fileChecksum uint16
}

func NewZHeader(story []byte) *ZHeader {
	zmem := new(ZHeader)
	zmem.configure(story)
	return zmem
}

func (zmem *ZHeader) configure(story []byte) {
	ByteAt := func(addr uint16) byte {
		return ReadZByte(story, addr)
	}

	WordAt := func(addr uint16) uint16 {
		return ReadZWord(story, addr)
	}

	zmem.version = ByteAt(0)

	if zmem.version > 3 {
		panic("versions > 3 are not supported!")
	}

	zmem.config = ByteAt(1)
	zmem.release = WordAt(2)

	zmem.highStart = WordAt(4)

	zmem.pc = WordAt(6)

	zmem.dictPos = WordAt(8)
	zmem.objTblPos = WordAt(0xA)
	zmem.globalsPos = WordAt(0xC)

	zmem.dynMemSize = WordAt(0xE)

	if zmem.dynMemSize < 64 {
		panic("dynamic size cannot be < 64 bytes")
	}

	if zmem.highStart < zmem.dynMemSize {
		panic("invalid story: high memory must not overlap dynamic memory")
	}

	// who cares if dynMemSize + staticMemorySize(min(0xFFFF, fileSize)) < 64KB ?

	for i := 0; i < SerialSize; i++ {
		zmem.serial[i] = ByteAt(uint16(0x12 + i))
	}

	zmem.abbrTblPos = WordAt(0x18)

	// v3
	zmem.fileLength = uint64(WordAt(0x1A)) * 2

	// v3 max file length 128K
	if zmem.fileLength > 128*1024 {
		panic("story file too big!")
	}

	zmem.fileChecksum = WordAt(0x1C)
}

func (zmem *ZHeader) Dump() {
	fmt.Printf("version: %d\n", zmem.version)

	if zmem.config&0x01 == 0x01 {
		fmt.Println("display hours:min")
	} else {
		fmt.Println("display score/turns")
	}

	fmt.Printf("release: %d\n", zmem.release)
	fmt.Printf("high memory start: %X\n", zmem.highStart)
	fmt.Printf("program counter: %X\n", zmem.pc)
	fmt.Printf("dictionary pos: %X\n", zmem.dictPos)
	fmt.Printf("object table pos: %X\n", zmem.objTblPos)
	fmt.Printf("global variables pos: %X\n", zmem.globalsPos)
	fmt.Printf("size of dynamic memory: %X\n", zmem.dynMemSize)
	fmt.Printf("serial number: %c%c%c%c%c%c\n", zmem.serial[0], zmem.serial[1], zmem.serial[2], zmem.serial[3], zmem.serial[4], zmem.serial[5])
	fmt.Printf("abbreviations table pos: %X\n", zmem.abbrTblPos)
	fmt.Printf("fileLength: %X\nfileChecksum: %X\n", zmem.fileLength, zmem.fileChecksum)
}
