package buffer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/susji/ked/gapbuffer"
)

type Buffer struct {
	lines []*gapbuffer.GapBuffer
	file  *os.File
}

func New(rawlines [][]rune) *Buffer {
	ret := &Buffer{}
	ret.lines = []*gapbuffer.GapBuffer{}
	for _, rawline := range rawlines {
		ret.lines = append(ret.lines, gapbuffer.NewFrom(rawline))
	}
	return ret
}

func NewFromFile(f *os.File) (*Buffer, error) {
	lines := []*gapbuffer.GapBuffer{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, gapbuffer.NewFrom([]rune(string(s.Bytes()))))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return &Buffer{
		lines: lines,
		file:  f,
	}, nil
}

func (b *Buffer) File() *os.File {
	return b.file
}

func (b *Buffer) Save() {
	if b.file == nil {
		panic("Save: no file backing this buffer")
	}
	panic("NOTIMPLEMENTED")
}

func (b *Buffer) NewLine(pos int) {
	if pos < 0 || len(b.lines) < pos {
		panic(fmt.Sprintf("NewLine: invalid pos=%d", pos))
	}

	left := b.lines[:pos]
	right := b.lines[pos:]
	b.lines = append(left, gapbuffer.New(gapbuffer.DEFAULTSZ))
	b.lines = append(b.lines, right...)
}

func (b *Buffer) DeleteLine(pos int) {
	if pos < 0 || len(b.lines) < pos {
		panic(fmt.Sprintf("DeleteLine: invalid pos=%d", pos))
	}
	left := b.lines[:pos]
	right := b.lines[pos+1:]
	b.lines = append(left, right...)

}

func (b *Buffer) Lines() []*gapbuffer.GapBuffer {
	return b.lines
}
