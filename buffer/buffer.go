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

func (b *Buffer) SetFilepath(filepath string) {
	b.filepath = filepath
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

func (b *Buffer) GetLine(lineno int) []rune {
	if lineno < 0 || len(b.lines) < lineno {
		panic(fmt.Sprintf("GetLine: invalid lineno=%d", lineno))
	}
	return b.lines[lineno].Get()
}

func (b *Buffer) LineLength(lineno int) int {
	if lineno < 0 || len(b.lines) < lineno {
		panic(fmt.Sprintf("GetLine: invalid lineno=%d", lineno))
	}
	return b.lines[lineno].Length()
}

func (b *Buffer) InsertRune(lineno, col int, r rune) int {
	b.lines[lineno].SetCursor(col)
	b.lines[lineno].Insert([]rune{r})
	return col + 1
}

func (b *Buffer) InsertRunes(lineno, col int, rs []rune) int {
	b.lines[lineno].SetCursor(col)
	b.lines[lineno].Insert(rs)
	return col + len(rs)
}

func (b *Buffer) InsertLinefeed(lineno, col int) (newlineno int, newcol int) {
	line := b.lines[lineno].Get()
	oldline := line[:col]
	newline := line[col:]

	b.lines[lineno].Clear().Insert(oldline)
	b.NewLine(lineno + 1).Insert(newline)

	return lineno + 1, 0
}

func (b *Buffer) Backspace(lineno, col int) (newlineno int, newcol int) {
	line := b.lines[lineno]
	linerunes := line.Get()
	if col == 0 && lineno > 0 {
		b.DeleteLine(lineno)
		if lineno > 0 {
			lineup := b.lines[lineno-1]
			lineuprunes := lineup.Get()
			lineup.SetCursor(len(lineuprunes))
			lineup.Insert(linerunes[col:])
			lineno--
			col = len(lineuprunes)
		}
		return lineno, col
	} else if col == 0 {
		return lineno, col
	}
	line.SetCursor(col)
	line.Delete()
	col--
	return lineno, col
}

func (b *Buffer) Lines() int {
	return len(b.lines)
}

func (b *Buffer) DeleteLineContent(lineno, col int) (newlineno int) {
	if b.LineLength(lineno) == 0 && b.Lines() > 1 {
		b.DeleteLine(lineno)
		if lineno == b.Lines() {
			return lineno - 1
		}
		return lineno
	}

	for b.LineLength(lineno) > col {
		b.Backspace(lineno, col+1)
	}
	return lineno
}
