package buffer

import "github.com/susji/ked/gapbuffer"

type Buffer struct {
	lines []*gapbuffer.GapBuffer
}

func New(rawlines [][]rune) *Buffer {
	ret := &Buffer{}
	for _, rawline := range rawlines {
		ret.lines = append(ret.lines, gapbuffer.NewFrom(rawline))
	}
	return ret
}

func (b *Buffer) Lines() []*gapbuffer.GapBuffer {
	return b.lines
}
