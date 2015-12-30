package gork

import (
	"strings"
	"unicode"
)

func SkipWhites(s string, offset int) int {
	for offset < len(s) && unicode.IsSpace(rune(s[offset])) {
		offset++
	}
	offset--
	return offset
}

func SplitSentence(sentence string, wordsep string) []string {
	ret := []string{}

	sentence = strings.Trim(sentence, " \r\n")

	wordStart := 0
	wordEnd := 0

	for i := 0; i < len(sentence); i++ {
		if unicode.IsSpace(rune(sentence[i])) {
			ret = append(ret, sentence[wordStart:wordEnd+1])

			i = SkipWhites(sentence, i)
			wordStart = i + 1
			wordEnd = wordStart
		} else {
			j := strings.IndexByte(wordsep, sentence[i])
			if j < 0 {
				wordEnd = i
			} else {
				if wordStart != wordEnd || wordStart != i {
					ret = append(ret, sentence[wordStart:wordEnd+1])
				}
				ret = append(ret, sentence[i:i+1])

				i++
				if i < len(sentence) {
					i = SkipWhites(sentence, i)
				}
				wordStart = i + 1
				wordEnd = wordStart

			}
		}
	}

	if wordStart < len(sentence) {
		ret = append(ret, sentence[wordStart:])
	}

	return ret
}
