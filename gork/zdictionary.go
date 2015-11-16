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

func NewZDictionary(story *ZStory, dictPos uint16, abbrTblPos uint16) *ZDictionary {
	zdict := new(ZDictionary)

	story.pos = dictPos
	n := story.ReadByte()

	for i := uint8(0); i < n; i++ {
		wordSep := story.ReadByte()
		zdict.wordSeparators = append(zdict.wordSeparators, wordSep)
	}

	zdict.entrySize = story.ReadByte()

	entryCount := story.ReadWord()

	for i := uint16(0); i < entryCount; i++ {
		word := DecodeZString(story, story.pos, abbrTblPos)
		zdict.words = append(zdict.words, word)
		story.pos += uint16(zdict.entrySize)
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
