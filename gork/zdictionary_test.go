package gork

import (
	"fmt"
	"testing"
)

var dictBuf []byte = []byte{
	3,
	'.', ',', '"',
	4,
	0, 2,

	// sorted order
	0x23, 0xC8, 0xC6, 0x95,
	0x7E, 0x97, 0xC0, 0xA5,
}

var dictExpected ZDictionary = ZDictionary{
	wordSeparators: []byte{'.', ',', '"'},
	entrySize:      4,
	entriesPos:     7,

	// sorted order
	words: []string{
		"cyclop",
		"zork",
	},
}

func TestZDictionary(t *testing.T) {
	mem := ZMemory(dictBuf)

	res := NewZDictionary(&mem, &ZHeader{dictPos: 0})

	for i, sep := range dictExpected.wordSeparators {
		if res.wordSeparators[i] != sep {
			t.Fail()
		}
	}

	if dictExpected.entrySize != res.entrySize ||
		dictExpected.entriesPos != res.entriesPos {
		t.Fail()
	}

	for i := range dictExpected.words {
		if dictExpected.words[i] != res.words[i] {
			t.Fail()
		}
	}

}

func TestZDictionarySearch(t *testing.T) {
	mem := ZMemory(dictBuf)

	dict := NewZDictionary(&mem, &ZHeader{dictPos: 0})

	randomData := []string{
		"42 is the answer",
		"73 is Chuck Norris of numbers",
		"golang",
	}

	fmt.Println(dict.words)
	for i, w := range dictExpected.words {
		pos := dictExpected.entriesPos + uint32(i)*uint32(dictExpected.entrySize)

		if dict.Search(w) != uint16(pos) {
			t.Fail()
		}
	}

	for _, d := range randomData {
		if dict.Search(d) != 0 {
			t.Fail()
		}
	}
}
