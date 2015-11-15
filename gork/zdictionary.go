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

func NewZDictionary(story []byte, dictPos uint16, abbrTblPos uint16) *ZDictionary {
	zdict := new(ZDictionary)

	n := ReadZByte(story, dictPos)

	addr := dictPos

	for i := uint8(0); i < n; i++ {
		addr++
		wordSep := ReadZByte(story, addr)
		zdict.wordSeparators = append(zdict.wordSeparators, wordSep)
	}
	addr++

	zdict.entrySize = ReadZByte(story, addr)
	addr++

	entryCount := ReadZWord(story, addr)
	addr += 2

	for i := uint16(0); i < entryCount; i++ {
		word := DecodeZString(story, addr, abbrTblPos)
		zdict.words = append(zdict.words, word)
		addr += uint16(zdict.entrySize)
	}

	return zdict
}

func (zdict *ZDictionary) String() string {
	ret := "\n    **** Dictionary ****\n\n"
	ret += fmt.Sprintf("  Word separators = \"%s\"\n", zdict.wordSeparators)
	ret += fmt.Sprintf("  Word count = %d, word size = %d\n\n", len(zdict.words), zdict.entrySize)

	for i, word := range zdict.words {
		ret += fmt.Sprintf("  [%4d] %s\n", i, word)
	}

	return ret
}
