package gork

import "testing"

var abbrsBuffer [][]byte = [][]byte{
	[]byte{0x7E, 0x97, 0xC0, 0xA5},
	[]byte{0x23, 0xC8, 0xC6, 0x95},
}

var abbrsExpected []string = []string{
	"zork",
	"cyclop",
}

func createMemory() *ZMemory {
	retLen := abbrCount * 2

	abbrsBufferLen := len(abbrsBuffer)
	abbrsPos := make([]uint32, abbrsBufferLen)

	abbrsPos[0] = uint32(retLen)
	for i := 1; i < abbrsBufferLen; i++ {
		abbrsPos[i] = abbrsPos[i-1] + uint32(len(abbrsBuffer[i%abbrsBufferLen]))
	}

	ret := make([]byte, retLen)

	for i := byte(0); i < abbrCount; i++ {

		// big endian
		j := i * 2
		pos := abbrsPos[int(i)%abbrsBufferLen] / 2
		ret[j] = byte(pos >> 8)
		ret[j+1] = byte(pos)
	}

	for _, abbr := range abbrsBuffer {
		ret = append(ret, abbr...)
	}

	tmp := ZMemory(ret)
	return &tmp
}

func TestGetAbbreviations(t *testing.T) {
	zmem := createMemory()
	abbrs := GetAbbreviations(zmem, &ZHeader{abbrTblPos: 0})

	for i := range abbrsExpected {
		if abbrsExpected[i] != abbrs[i] {
			t.Fail()
		}
	}
}
