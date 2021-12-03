package buffer

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/susji/ked/gapbuffer"
)

type Buffer struct {
	lines    []*gapbuffer.GapBuffer
	filepath string
}

func New(rawlines [][]rune) *Buffer {
	ret := &Buffer{}
	ret.lines = []*gapbuffer.GapBuffer{}
	for _, rawline := range rawlines {
		ret.lines = append(ret.lines, gapbuffer.NewFrom(rawline))
	}
	return ret
}

func NewFromReader(filepath string, r io.Reader) (*Buffer, error) {
	lines := []*gapbuffer.GapBuffer{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		lines = append(lines, gapbuffer.NewFrom([]rune(string(s.Bytes()))))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		lines = append(lines, gapbuffer.New(gapbuffer.DEFAULTSZ))
	}
	return &Buffer{
		lines:    lines,
		filepath: filepath,
	}, nil
}

func (b *Buffer) Filepath() string {
	return b.filepath
}

func (b *Buffer) Save() error {
	if len(b.filepath) == 0 {
		panic("Save: no file backing this buffer")
	}
	data := []byte{}
	for _, gb := range b.lines {
		linedata := []byte(string(gb.Get()))
		linedata = append(linedata, '\n')
		data = append(data, linedata...)
	}
	return os.WriteFile(b.filepath, data, 0644)
}

func (b *Buffer) NewLine(pos int) *gapbuffer.GapBuffer {
	if pos < 0 || len(b.lines) < pos {
		panic(fmt.Sprintf("NewLine: invalid pos=%d", pos))
	}

	left := b.lines[:pos]
	right := b.lines[pos:]
	newline := gapbuffer.New(gapbuffer.DEFAULTSZ)

	newlines := make([]*gapbuffer.GapBuffer, len(left)+len(right)+1)

	copy(newlines, left)
	newlines[len(left)] = newline
	copy(newlines[len(left)+1:], right)
	b.lines = newlines

	//log.Printf("[NewLine=%d] left=%q  right=%q => %q\n", pos, left, right, b.lines)
	return newline
}

func (b *Buffer) DeleteLine(pos int) {
	if pos < 0 || len(b.lines) < pos {
		panic(fmt.Sprintf("DeleteLine: invalid pos=%d", pos))
	}
	left := b.lines[:pos]
	right := b.lines[pos+1:]
	//log.Printf("[DeleteLine=%d] left=%q  right=%q\n", pos, left, right)
	b.lines = append(left, right...)
}

func (b *Buffer) GetLine(pos int) *gapbuffer.GapBuffer {
	if pos < 0 || len(b.lines) < pos {
		panic(fmt.Sprintf("GetLine: invalid pos=%d", pos))
	}
	return b.lines[pos]
}

func (b *Buffer) Lines() int {
	return len(b.lines)
}
