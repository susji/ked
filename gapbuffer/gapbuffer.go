package gapbuffer

import (
	"encoding/hex"
	"fmt"
	"os"
)

const DEFAULTSZ = 64

type GapBuffer struct {
	buf []rune
	// pre means cursor position in buf
	pre int
	// post means position in buf after cursor and gap
	post int
}

func New(sz int) *GapBuffer {
	return &GapBuffer{
		buf:  make([]rune, sz),
		pre:  0,
		post: sz,
	}
}

func hexdump(what []rune) {
	d := hex.Dumper(os.Stderr)
	d.Write([]byte(string(what)))
}

func debug(f string, va ...interface{}) {
	fmt.Fprintf(os.Stderr, f+"\n", va...)
}

func (gb *GapBuffer) Cursor() int {
	return gb.pre
}

func (gb *GapBuffer) SetCursor(cursor int) {
	debug("(SetCursor before) pre=%d  post=%d  {%d <- %d}", gb.pre, gb.post, cursor, gb.pre)
	if cursor > gb.Length() {
		panic("cursor > gb.Length")
	}
	//
	// Moving the cursor (gb.pre) of a GapBuffer looks like this:
	//
	//
	//
	// /--------------1234/    /abcdefghijklmn|
	//                 pre^     ^post
	//
	//
	// /--------------1234abcde/   /efghijklmn|
	//                      pre^   ^post
	//
	// So there are two cases we may have here:
	//
	//   #1  Gap moves left
	//   #2  Gap moves right
	//
	// In both cases, the
	//
	newpre := cursor
	var dir, moves int
	if newpre < gb.pre {
		dir = -1
		moves = gb.pre - newpre
	} else {
		dir = 1
		moves = newpre - gb.pre
	}

	for i := 0; i < moves; i++ {
		to := gb.post - 1
		from := gb.pre - 1
		debug("{%02d} %d <- %d", i, to, from)
		gb.buf[to] = gb.buf[from]
		gb.pre += dir
		gb.post += dir
	}
	debug("(SetCursor afterwards) pre=%d  post=%d", gb.pre, gb.post)
	hexdump(gb.buf)
}

func (gb *GapBuffer) Length() int {
	return len(gb.buf) - (gb.post - gb.pre)
}

func (gb *GapBuffer) gaplen() int {
	return gb.post - gb.pre
}

func (gb *GapBuffer) gapgrow(atleast int) {
	//
	// Increasing the length of a GapBuffer looks like this:
	//
	//                         012345678abcdef
	// /----------------/ gap /---------------|
	//                 pre  oldpost
	//                                  0123456789abcdef
	// /----------------/ gapgapgapgap /----------------|
	//                 pre          newpost
	//
	n := len(gb.buf) - gb.post
	gb.buf = append(gb.buf, make([]rune, atleast)...)

	srcstart := gb.post
	srcend := gb.post + n
	dststart := gb.post + atleast
	dstend := gb.post + atleast + n
	copy(gb.buf[dststart:dstend], gb.buf[srcstart:srcend])
	gb.post += atleast
}

func (gb *GapBuffer) Insert(what []rune) {
	//
	// Inserting into a GapBuffer looks like this:
	//
	//                           what
	//                         -======-
	//
	// /-----------------------/ gapgapgap /----------------|
	//                        pre
	//
	if gb.gaplen() <= len(what) {
		gb.gapgrow(len(what))
	}
	copy(gb.buf[gb.pre:], what)
	gb.pre += len(what)
	hexdump(gb.buf)
}

func (gb *GapBuffer) Delete() {
	if gb.pre == 0 {
		panic("Delete: pre == 0")
	}
	gb.pre--
}

func (gb *GapBuffer) Get() []rune {
	debug("(Get) pre=%d  post=%d", gb.pre, gb.post)
	ret := make([]rune, len(gb.buf)-gb.gaplen())
	copy(ret, gb.buf[:gb.pre])
	copy(ret[gb.pre:], gb.buf[gb.post:])
	return ret
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
