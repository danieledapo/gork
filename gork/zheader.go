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

func NewZHeader(story *ZStory) *ZHeader {
	zmem := new(ZHeader)
	zmem.configure(story)
	return zmem
}

func (zmem *ZHeader) configure(story *ZStory) {
	story.pos = 0

	zmem.version = story.ReadByte()

	if zmem.version > 3 {
		panic("versions > 3 are not supported!")
	}

	zmem.config = story.ReadByte()
	zmem.release = story.ReadWord()

	zmem.highStart = story.ReadWord()

	zmem.pc = story.ReadWord()

	zmem.dictPos = story.ReadWord()
	zmem.objTblPos = story.ReadWord()
	zmem.globalsPos = story.ReadWord()

	zmem.dynMemSize = story.ReadWord()

	if zmem.dynMemSize < 64 {
		panic("dynamic size cannot be < 64 bytes")
	}

	if zmem.highStart < zmem.dynMemSize {
		panic("invalid story: high memory must not overlap dynamic memory")
	}

	// who cares if dynMemSize + staticMemorySize(min(0xFFFF, fileSize)) < 64KB ?

	story.pos = 0x12
	for i := 0; i < SerialSize; i++ {
		zmem.serial[i] = story.ReadByte()
	}

	zmem.abbrTblPos = story.ReadWord()

	// v3
	zmem.fileLength = uint64(story.ReadWord()) * 2

	// v3 max file length 128K
	if zmem.fileLength > 128*1024 {
		panic("story file too big!")
	}

	zmem.fileChecksum = story.ReadWord()
}

func (zmem *ZHeader) String() string {
	ret := "\n    **** Story file header ****\n\n"
	ret += fmt.Sprintf("  Z-code version:           %d\n", zmem.version)

	ret += fmt.Sprint("  Interpreter flags:        ")
	if zmem.config&0x01 == 0x01 {
		ret += fmt.Sprintln("Display hours:min")
	} else {
		ret += fmt.Sprintln("Display score/turns")
	}

	ret += fmt.Sprintf("  Release number:           %d\n", zmem.release)
	ret += fmt.Sprintf("  Size of resident memory:  %04x\n", zmem.highStart)
	ret += fmt.Sprintf("  Start PC:                 %04x\n", zmem.pc)
	ret += fmt.Sprintf("  Dictionary address:       %04x\n", zmem.dictPos)
	ret += fmt.Sprintf("  Object table address:     %04x\n", zmem.objTblPos)
	ret += fmt.Sprintf("  Global variables address: %04x\n", zmem.globalsPos)
	ret += fmt.Sprintf("  Size of dynamic memory:   %04x\n", zmem.dynMemSize)
	ret += fmt.Sprintf("  Serial number:            %c%c%c%c%c%c\n", zmem.serial[0], zmem.serial[1], zmem.serial[2], zmem.serial[3], zmem.serial[4], zmem.serial[5])
	ret += fmt.Sprintf("  Abbreviations address:    %04x\n", zmem.abbrTblPos)
	ret += fmt.Sprintf("  File size:                %05x\n", zmem.fileLength)
	ret += fmt.Sprintf("  Checksum:                 %04x\n", zmem.fileChecksum)

	return ret
}
