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

func NewZHeader(mem *ZMemory) *ZHeader {
	zmem := new(ZHeader)
	zmem.configure(mem)
	return zmem
}

func (header *ZHeader) configure(mem *ZMemory) {
	seq := mem.GetSequential(0)

	header.version = seq.ReadByte()

	if header.version > 3 {
		panic("versions > 3 are not supported!")
	}

	header.config = seq.ReadByte()
	header.release = seq.ReadWord()

	header.highStart = seq.ReadWord()

	header.pc = seq.ReadWord()

	header.dictPos = seq.ReadWord()
	header.objTblPos = seq.ReadWord()
	header.globalsPos = seq.ReadWord()

	header.dynMemSize = seq.ReadWord()

	if header.dynMemSize < 64 {
		panic("dynamic size cannot be < 64 bytes")
	}

	if header.highStart < header.dynMemSize {
		panic("invalid mem: high memory must not overlap dynamic memory")
	}

	// who cares if dynMemSize + staticMemorySize(min(0xFFFF, fileSize)) < 64KB ?

	seq.pos = 0x12
	for i := 0; i < SerialSize; i++ {
		header.serial[i] = seq.ReadByte()
	}

	header.abbrTblPos = seq.ReadWord()

	// v3
	header.fileLength = uint64(seq.ReadWord()) * 2

	// v3 max file length 128K
	if header.fileLength > 128*1024 {
		panic("mem file too big!")
	}

	header.fileChecksum = seq.ReadWord()
}

func (header *ZHeader) String() string {
	ret := "\n    **** Story file header ****\n\n"
	ret += fmt.Sprintf("  Z-code version:           %d\n", header.version)

	ret += fmt.Sprint("  Interpreter flags:        ")
	if header.config&0x01 == 0x01 {
		ret += fmt.Sprintln("Display hours:min")
	} else {
		ret += fmt.Sprintln("Display score/turns")
	}

	ret += fmt.Sprintf("  Release number:           %d\n", header.release)
	ret += fmt.Sprintf("  Size of resident memory:  %04x\n", header.highStart)
	ret += fmt.Sprintf("  Start PC:                 %04x\n", header.pc)
	ret += fmt.Sprintf("  Dictionary address:       %04x\n", header.dictPos)
	ret += fmt.Sprintf("  Object table address:     %04x\n", header.objTblPos)
	ret += fmt.Sprintf("  Global variables address: %04x\n", header.globalsPos)
	ret += fmt.Sprintf("  Size of dynamic memory:   %04x\n", header.dynMemSize)
	ret += fmt.Sprintf("  Serial number:            %c%c%c%c%c%c\n", header.serial[0], header.serial[1], header.serial[2], header.serial[3], header.serial[4], header.serial[5])
	ret += fmt.Sprintf("  Abbreviations address:    %04x\n", header.abbrTblPos)
	ret += fmt.Sprintf("  File size:                %05x\n", header.fileLength)
	ret += fmt.Sprintf("  Checksum:                 %04x\n", header.fileChecksum)

	return ret
}
