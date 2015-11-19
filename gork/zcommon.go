package gork

import (
	"fmt"
)

// v3
var Alphabets = [3]string{
	"abcdefghijklmnopqrstuvwxyz",
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	" \n0123456789.,!?_#'\"/\\-:()",
}

func GetAbbreviations(story *ZStory, header *ZHeader) []string {
	// v3 3 tables * 32 entries each
	const abbrCount = 32 * 3

	story.pos = header.abbrTblPos

	ret := []string{}

	for i := uint16(0); i < abbrCount; i++ {
		addr := story.ReadWord() * 2
		tmpPos := story.pos
		ret = append(ret, DecodeZString(story, addr, header))
		story.pos = tmpPos
	}

	return ret
}

func DecodeZString(story *ZStory, addr uint16, header *ZHeader) string {
	// v3

	ret := ""
	data := uint16(0)
	code := uint16(0)

	alphabet := uint8(0)
	shiftLock := uint8(0)

	synonimFlag := false
	synonim := uint16(0)

	// 0 not ascii
	// 1 first part
	// 2 last part
	asciiPart := uint8(0)
	asciiFirstPart := uint16(0)

	// save current position
	oldPos := story.pos
	story.pos = addr

	for data&0x8000 == 0 {
		data = story.ReadWord()

		for i := 10; i >= 0; i -= 5 {
			code = (data >> uint8(i)) & 0x1F

			if synonimFlag {
				synonimFlag = false
				synonim = (synonim - 1) * 64

				oldPos := story.pos
				story.pos = header.abbrTblPos + synonim + code*2
				tmpAddr := story.ReadWord() * 2
				ret += DecodeZString(story, tmpAddr, header)
				story.pos = oldPos

				alphabet = shiftLock
			} else if asciiPart > 0 {
				tmp := asciiPart
				asciiPart++
				if tmp == 1 {
					asciiFirstPart = code << 5
				} else {
					asciiPart = 0
					ret += string(asciiFirstPart | code)
				}
			} else if code > 5 {
				code -= 6

				if alphabet == 2 && code == 0 {
					asciiPart = 1
				} else if alphabet == 2 && code == 1 {
					ret += "\n"
				} else {
					ret += string(Alphabets[alphabet][code])
				}
				alphabet = shiftLock
			} else if code == 0 {
				ret += " "
			} else if code < 4 {
				synonimFlag = true
				synonim = code
			} else {
				alphabet = uint8(code - 3)
				shiftLock = 0
			}
		}
	}

	// restore old position
	story.pos = oldPos
	return ret
}

func DumpAbbreviations(story *ZStory, header *ZHeader) {
	fmt.Print("\n    **** Abbreviations ****\n\n")

	abbrs := GetAbbreviations(story, header)

	if len(abbrs) == 0 {
		fmt.Printf("  No abbreviation information.\n")
		return
	}

	for i, abbr := range abbrs {
		fmt.Printf("  [%2d] \"%s\"\n", i, abbr)
	}
}

func PackedAddress(addr uint16) uint16 {
	// v3
	return addr * 2
}

func IsPackedAddress(addr uint16) bool {
	// v3
	return addr%2 == 0
}
