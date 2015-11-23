package gork

import (
	"fmt"
)

type ZDictionary struct {
	wordSeparators []byte
	entrySize      uint8
	words          []string
	// ignore words data, it looks like they are useless to interpreters
}

func NewZDictionary(mem *ZMemory, header *ZHeader) *ZDictionary {
	zdict := new(ZDictionary)

	seq := mem.GetSequential(header.dictPos)

	n := seq.ReadByte()

	for i := uint8(0); i < n; i++ {
		wordSep := seq.ReadByte()
		zdict.wordSeparators = append(zdict.wordSeparators, wordSep)
	}

	zdict.entrySize = seq.ReadByte()

	entryCount := seq.ReadWord()

	for i := uint16(0); i < entryCount; i++ {
		word := mem.DecodeZStringAt(seq.pos, header)
		zdict.words = append(zdict.words, word)
		seq.pos += uint16(zdict.entrySize)
	}

	return zdict
}

func (zdict *ZDictionary) String() string {
	ret := "\n    **** Dictionary ****\n\n"
	ret += fmt.Sprintf("  Word separators = \"%s\"\n", zdict.wordSeparators)
	ret += fmt.Sprintf("  Word count = %d, word size = %d\n\n", len(zdict.words), zdict.entrySize)

	for i, word := range zdict.words {
		ret += fmt.Sprintf("  [%4d] %s\n", i+1, word)
	}

	return ret
}
