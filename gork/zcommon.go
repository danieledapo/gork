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

func ReadZByte(buf []byte, addr uint16) byte {
	return buf[addr]
}

func ReadZWord(buf []byte, addr uint16) uint16 {
	// Big Endian
	return (uint16(buf[addr]) << 8) | (uint16(buf[addr+1]))
}

func ReadUint32(buf []byte, addr uint16) uint32 {
	// Big Endian
	return (uint32(buf[addr]) << 24) | (uint32(buf[addr+1]) << 16) |
		(uint32(buf[addr+2]) << 8) | uint32(buf[addr+3])
}

func GetAbbreviations(story []byte, abbrTblPos uint16) []string {
	// v3 3 tables * 32 entries each
	const abbrCount = 32 * 3

	ret := []string{}

	for i := uint16(0); i < abbrCount; i++ {
		addr := ReadZWord(story, abbrTblPos+i*2) * 2
		ret = append(ret, DecodeZString(story, addr, abbrTblPos))
	}

	return ret
}

func DecodeZString(story []byte, addr uint16, abbrTblPos uint16) string {
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

	offset := uint16(0)

	for data&0x8000 == 0 {
		data = ReadZWord(story, addr+offset)
		offset += 2

		for i := 10; i >= 0; i -= 5 {
			code = (data >> uint8(i)) & 0x1F

			if synonimFlag {
				synonimFlag = false
				synonim = (synonim - 1) * 64
				pos := abbrTblPos + synonim + code*2
				tmpAddr := ReadZWord(story, pos) * 2
				ret += DecodeZString(story, tmpAddr, abbrTblPos)
				alphabet = shiftLock
			} else if asciiPart > 0 {
				if asciiPart++; asciiPart == 1 {
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

	return ret
}

func DumpAbbreviations(story []byte, abbrTblPos uint16) {
	fmt.Print("\n    **** Abbreviations ****\n\n")

	abbrs := GetAbbreviations(story, abbrTblPos)

	if len(abbrs) == 0 {
		fmt.Printf("  No abbreviation information.\n")
		return
	}

	for i, abbr := range abbrs {
		fmt.Printf("  [%2d] \"%s\"\n", i, abbr)
	}
}
