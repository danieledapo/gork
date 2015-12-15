package gork

import "testing"

var headerBuf []byte = []byte{
	3,    // version
	1,    // config (hours:min vs score/turns)
	0, 1, // release
	0x4E, 0x37, // high memory start position (aka resident memory size)
	0x4F, 0x05, // initial PC
	0x3B, 0x21, // dictionary address
	0x02, 0xB0, // object table address
	0x22, 0x71, // global variables address
	0x2E, 0x53, // size of dynamic memory
	0xFF, 0xFF, // empty
	0, 0, 0, 0, 0, 1, // serial
	0x01, 0xF0, // abbreviations table address
	0xA5, 0xC6, // file size
	0xa1, 0x29, // file checksum
}

func ZHeaderConfigureTest(t *testing.T) {
	mem := ZMemory(headerBuf)
	header := NewZHeader(&mem)

	shouldPass := header.config == 1 &&
		header.version == 3 &&
		header.release == 1 &&
		header.highStart == 0x4E37 &&
		header.pc == 0x4F05 &&
		header.dictPos == 0x3B21 &&
		header.objTblPos == 0x02B0 &&
		header.globalsPos == 0x2271 &&
		header.dynMemSize == 0x2E53 &&
		header.serial == [SerialSize]byte{0, 0, 0, 0, 0, 1} &&
		header.abbrTblPos == 0x01F0 &&
		header.fileLength == uint64(0xA5C6)*2 &&
		header.fileChecksum == 0xa129

	if !shouldPass {
		t.Fail()
	}
}
