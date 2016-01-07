package gork

// v3 3 tables * 32 entries each
const abbrCount = 32 * 3

func GetAbbreviations(mem *ZMemory, header *ZHeader) []string {

	seq := mem.GetSequential(uint32(header.abbrTblPos))

	ret := []string{}

	for i := uint16(0); i < abbrCount; i++ {
		addr := uint32(seq.ReadWord()) * 2
		ret = append(ret, mem.DecodeZStringAt(addr, header))
	}

	return ret
}
