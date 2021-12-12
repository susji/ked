package buffer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/susji/ked/config"
	"github.com/susji/ked/gapbuffer"
)

type Buffer struct {
	lines    []*gapbuffer.GapBuffer
	filepath string
	mods     []*modification
}

func New(rawlines [][]rune) *Buffer {
	ret := &Buffer{}
	ret.lines = []*gapbuffer.GapBuffer{}
	if len(rawlines) == 0 {
		rawlines = [][]rune{[]rune("")}
	}
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

func (b *Buffer) UndoModification() *ActionResult {
	if len(b.mods) == 0 {
		return nil
	}
restart:
	n := len(b.mods) - 1
	mod := b.mods[n]
	b.mods = b.mods[:n]
	log.Printf("[UndoModification]: %+v\n", mod)
	switch mod.kind {
	case MOD_INSERTRUNES:
		data := mod.data.([]rune)
		for i := 0; i < len(data); i++ {
			b.lines[mod.lineno].SetCursor(mod.col + 1).Delete()
		}
	case MOD_LINEFEED:
		lineno := mod.lineno

		newline := gapbuffer.New(gapbuffer.DEFAULTSZ)
		newline.Insert(b.lines[lineno].Get())
		newline.SetCursor(newline.Length())
		newline.Insert(b.lines[lineno+1].Get())
		left := b.lines[:lineno]
		right := b.lines[lineno+1:]
		b.lines = append(left, right...)
		b.lines[lineno] = newline
	case MOD_DELETERUNES:
		lineno := mod.lineno
		col := mod.col
		b.lines[lineno].SetCursor(col).Insert(mod.data.([]rune))
	case MOD_MOVERUNES:
		lineno := mod.lineno
		data := mod.data.([]rune)
		for i := 0; i < len(data); i++ {
			b.lines[lineno].SetCursor(mod.col + 1).Delete()
		}
		b.lines[lineno+1].SetCursor(0).Insert(data)
	case MOD_DELETELINE:
		lineno := mod.lineno
		left := b.lines[:lineno]
		right := b.lines[lineno:]
		newline := gapbuffer.New(gapbuffer.DEFAULTSZ)
		newlines := make([]*gapbuffer.GapBuffer, len(left)+len(right)+1)
		copy(newlines, left)
		newlines[len(left)] = newline
		copy(newlines[len(left)+1:], right)
		b.lines = newlines
	}
	// Execute all sequential modifications of the same kind.
	if len(b.mods) > 0 && mod.kind == b.mods[len(b.mods)-1].kind {
		goto restart
	}
	return &ActionResult{Lineno: mod.lineno, Col: mod.col}
}

func (b *Buffer) modify(mod *modification) {
	log.Printf("[modify] %+v\n", mod)
	b.mods = append(b.mods, mod)
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

func (b *Buffer) deleteline(act *Action) ActionResult {
	lineno := act.lineno
	if lineno < 0 || len(b.lines) < lineno {
		panic(fmt.Sprintf("deleteline: invalid lineno=%d", lineno))
	}
	left := b.lines[:lineno]
	right := b.lines[lineno+1:]
	b.lines = append(left, right...)
	b.modify(&modification{
		kind:   MOD_DELETELINE,
		lineno: lineno,
	})
	if lineno >= b.Lines() {
		lineno = b.Lines() - 1
	}
	return ActionResult{Lineno: lineno, Col: 0}
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

func (b *Buffer) insertlinefeed(act *Action) ActionResult {
	lineno := act.lineno
	col := act.col
	line := b.lines[lineno].Get()
	oldline := line[:col]
	newline := line[col:]

	b.lines[lineno].Clear().Insert(oldline)
	b.NewLine(lineno + 1).Insert(newline)
	b.modify(&modification{
		kind:   MOD_LINEFEED,
		lineno: lineno,
		col:    col,
	})
	return ActionResult{Lineno: lineno + 1, Col: 0}
}

func (b *Buffer) backspace(act *Action) ActionResult {
	line := b.lines[act.lineno]
	linerunes := line.Get()
	lineno := act.lineno
	col := act.col
	if col == 0 && lineno > 0 {
		lineup := b.lines[lineno-1]
		lineuprunes := lineup.Get()

		b.modify(&modification{
			kind:   MOD_MOVERUNES,
			lineno: lineno - 1,
			col:    len(lineuprunes),
			data:   linerunes[col:],
		})
		b.lines[lineno-1].SetCursor(len(lineuprunes)).Insert(linerunes[col:])

		b.Perform(NewDelLine(lineno))
		lineno--
		col = len(lineuprunes)
		return ActionResult{Lineno: lineno, Col: col}
	} else if col == 0 {
		return ActionResult{Lineno: lineno, Col: col}
	}
	b.modify(&modification{
		kind:   MOD_DELETERUNES,
		lineno: lineno,
		col:    col - 1,
		data:   []rune{linerunes[col-1]},
	})
	line.SetCursor(col)
	line.Delete()
	col--
	return ActionResult{Lineno: lineno, Col: col}
}

func (b *Buffer) Lines() int {
	return len(b.lines)
}

func (b *Buffer) deletelinecontent(act *Action) ActionResult {
	lineno := act.lineno
	col := act.col
	if b.LineLength(lineno) == 0 {
		panic("deletelinecontent: got an empty line")
	}
	line := b.lines[lineno]
	b.modify(&modification{
		kind:   MOD_DELETERUNES,
		lineno: lineno,
		col:    col,
		data:   line.Get()[col:],
	})
	for b.LineLength(lineno) > col {
		line.SetCursor(col + 1)
		line.Delete()
	}
	return ActionResult{Lineno: lineno, Col: col}
}

type SearchLimit struct {
	StartLineno, StartCol int
	EndLineno, EndCol     int
}

func (b *Buffer) Search(term []rune) (lineno, col int) {
	limits := &SearchLimit{
		StartLineno: 0,
		StartCol:    0,
		EndLineno:   b.Lines() - 1,
		EndCol:      b.LineLength(b.Lines() - 1),
	}
	return b.SearchRange(term, limits)
}

func (b *Buffer) SearchRange(term []rune, limits *SearchLimit) (lineno, col int) {
	sterm := string(term)
	for lineno := limits.StartLineno; lineno <= limits.EndLineno; lineno++ {
		line := b.GetLine(lineno)
		if len(line) == 0 {
			continue
		}
		a, b := 0, len(line)
		if lineno == limits.StartLineno {
			a = limits.StartCol
		}
		if lineno == limits.EndLineno {
			b = limits.EndCol
		}
		line = line[a:b]
		s := strings.ToLower(string(line))
		col := strings.Index(s, sterm)
		//log.Printf("ZZZ: %q/%q -> %d\n", sterm, s, col)
		if col >= 0 {
			return lineno, col + a
		}
	}
	return -1, -1
}

func (b *Buffer) NextRune(lineno, col int) (rune, error) {
	line := b.lines[lineno].Get()
	if col+1 < len(line) {
		return line[col+1], nil
	}
	nextlineno := lineno + 1
	if nextlineno >= b.Lines() {
		return ' ', errors.New("no next line => no next rune")
	}
	return b.lines[nextlineno].Get()[0], nil

}

func (b *Buffer) PrevRune(lineno, col int) (rune, error) {
	line := b.lines[lineno].Get()
	if col > 0 {
		return line[col-1], nil
	}
	prevlineno := lineno - 1
	if prevlineno < 0 {
		return ' ', errors.New("no previous line => no previous rune")
	}
	prevline := b.lines[prevlineno].Get()
	return prevline[len(prevline)-1], nil
}

func (b *Buffer) Modified() bool {
	return len(b.mods) > 0
}

func (b *Buffer) JumpWord(lineno, col int, left bool) (newlineno, newcol int) {
	origlineno := lineno
	origcol := col

	if left {
		for lineno >= 0 && lineno < b.Lines() {
			line := b.GetLine(lineno)
			var i int
			for i = col - 1; i > 0; i-- {
				pr, _ := b.PrevRune(lineno, i)
				if strings.ContainsAny(string(pr), config.WORD_DELIMS) &&
					!strings.ContainsAny(string(line[i]), config.WORD_DELIMS) {
					return lineno, i
				}
			}
			if lineno == 0 {
				return lineno, 0
			}
			lineno--
			col = b.LineLength(lineno)
		}
	} else {
		for lineno >= 0 && lineno < b.Lines() {
			line := b.GetLine(lineno)
			for i := col; i <= len(line)-1; i++ {
				if strings.ContainsAny(string(line[i]), config.WORD_DELIMS) {
					// We consider end-of-line as a delimiter.
					if i == len(line)-1 {
						return lineno, i + 1
					}
					// Skip subsequent word delimiters.
					nr, _ := b.NextRune(lineno, i)
					if !strings.ContainsAny(string(nr), config.WORD_DELIMS) {
						return lineno, i + 1
					}
				}
			}
			lineno++
			col = 0
		}
	}
	return origlineno, origcol
}

func (b *Buffer) insertrune(act *Action) ActionResult {
	rs := act.data.([]rune)
	b.lines[act.lineno].SetCursor(act.col)
	b.lines[act.lineno].Insert(rs)
	b.modify(&modification{
		kind:   MOD_INSERTRUNES,
		lineno: act.lineno,
		col:    act.col,
		data:   rs,
	})
	return ActionResult{Lineno: act.lineno, Col: act.col + 1}
}

func (b *Buffer) delword(act *Action) ActionResult {
	lineno := act.lineno
	col := act.col

	// This is the suboptimal version where we just
	// leverage backspace() instead of doing string
	// search to handle the word-deletion in one go.

keepgoing:
	res := b.backspace(NewBackspace(lineno, col))
	lineno = res.Lineno
	col = res.Col
	linerunes := string(b.GetLine(lineno))
	if col > 0 && !strings.ContainsAny(string(linerunes[col-1]), config.WORD_DELIMS) {
		goto keepgoing
	}
	return res
}

func (b *Buffer) detabulate(act *Action) ActionResult {
	// This many spaces is enough for everyone.
	spaceprefix := []rune("                                                    ")
	lineno := act.lineno
	col := act.col
	line := b.lines[lineno]
	runes := line.Get()
	if len(runes) > 0 && runes[0] == '\t' {
		b.modify(&modification{
			kind:   MOD_DELETERUNES,
			lineno: lineno,
			col:    0,
			data:   []rune{'\t'},
		})
		line.SetCursor(1).Delete()
		if col > 0 {
			col--
		}
		return ActionResult{Lineno: lineno, Col: col}
	} else if len(runes) >= config.TABSZ &&
		reflect.DeepEqual(runes[:config.TABSZ], spaceprefix[:config.TABSZ]) {
		b.modify(&modification{
			kind:   MOD_DELETERUNES,
			lineno: lineno,
			col:    0,
			data:   []rune("    "),
		})
		for i := 0; i < config.TABSZ; i++ {
			line.SetCursor(1).Delete()
			if col > 0 {
				col--
			}
		}
	}
	return ActionResult{Lineno: lineno, Col: col}
}

// Perform is our action dispatch. This should be the only way
// for outsiders to generate changes in buffer contents. Here
// we also handle all the relevant book-keepping for undo.
func (b *Buffer) Perform(act *Action) ActionResult {
	dispatch := map[ActionKind]ActionFunc{
		ACT_RUNES:          b.insertrune,
		ACT_BACKSPACE:      b.backspace,
		ACT_LINEFEED:       b.insertlinefeed,
		ACT_DELLINECONTENT: b.deletelinecontent,
		ACT_DELLINE:        b.deleteline,
		ACT_DETABULATE:     b.detabulate,
		ACT_DELWORD:        b.delword,
	}
	return dispatch[act.kind](act)
}
