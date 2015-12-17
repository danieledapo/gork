package gork

import "testing"

var dictBuf []byte = []byte{
	3,
	'.', ',', '"',
	4,
	0, 2,
	0x7E, 0x97, 0xC0, 0xA5,
	0x23, 0xC8, 0xC6, 0x95,
}

var dictExpected ZDictionary = ZDictionary{
	wordSeparators: []byte{'.', ',', '"'},
	entrySize:      4,
	words: []string{
		"zork",
		"cyclop",
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

	if dictExpected.entrySize != res.entrySize {
		t.Fail()
	}

	for i := range dictExpected.words {
		if dictExpected.words[i] != res.words[i] {
			t.Fail()
		}
	}

}
