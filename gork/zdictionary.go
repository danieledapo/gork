package gork

import (
	"fmt"
	"sort"
)

type ZDictionary struct {
	wordSeparators []byte
	entrySize      uint8
	words          []string
	entriesPos     uint32
	// ignore words data, it looks like they are useless to interpreters
}

func NewZDictionary(mem *ZMemory, header *ZHeader) *ZDictionary {
	zdict := new(ZDictionary)

	seq := mem.GetSequential(uint32(header.dictPos))

	n := seq.ReadByte()

	for i := uint8(0); i < n; i++ {
		wordSep := seq.ReadByte()
		zdict.wordSeparators = append(zdict.wordSeparators, wordSep)
	}

	zdict.entrySize = seq.ReadByte()

	entryCount := seq.ReadWord()

	zdict.entriesPos = seq.pos

	for i := uint16(0); i < entryCount; i++ {
		word := mem.DecodeZStringAt(seq.pos, header)
		zdict.words = append(zdict.words, word)
		seq.pos += uint32(zdict.entrySize)
	}

	return zdict
}

func (dict *ZDictionary) Search(s string) uint16 {
	i := sort.SearchStrings(dict.words, s)

	if i < len(dict.words) && dict.words[i] == s {
		return uint16(dict.entriesPos + uint32(i)*uint32(dict.entrySize))
	}

	// not found
	return 0
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
