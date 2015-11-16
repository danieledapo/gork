package gork

import (
	"fmt"
)

type ZStory struct {
	buf []byte
	pos uint16
}

func NewZStory(story []byte) *ZStory {
	return &ZStory{story, 0}
}

func (zstory *ZStory) PeekByte() byte {
	return zstory.buf[zstory.pos]
}

func (zstory *ZStory) PeekWord() uint16 {
	// Big Endian
	return (uint16(zstory.buf[zstory.pos]) << 8) |
		(uint16(zstory.buf[zstory.pos+1]))
}

func (zstory *ZStory) PeekUInt32() uint32 {
	// Big Endian
	return (uint32(zstory.buf[zstory.pos]) << 24) |
		(uint32(zstory.buf[zstory.pos+1]) << 16) |
		(uint32(zstory.buf[zstory.pos+2]) << 8) |
		uint32(zstory.buf[zstory.pos+3])
}

func (zstory *ZStory) ReadByte() byte {
	tmp := zstory.PeekByte()
	zstory.pos++
	return tmp
}

func (zstory *ZStory) ReadWord() uint16 {
	tmp := zstory.PeekWord()
	zstory.pos += 2
	return tmp
}

func (zstory *ZStory) ReadUint32() uint32 {
	tmp := zstory.PeekUInt32()
	zstory.pos += 4
	return tmp
}

func (zstory *ZStory) String() string {
	return fmt.Sprintf("buf: %s pos: %d", zstory.buf, zstory.pos)
}
