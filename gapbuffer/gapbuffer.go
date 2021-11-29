package gapbuffer

import (
	"encoding/hex"
	"fmt"
	"os"
)

const initialsize = 64

type GapBuffer struct {
	buf []rune
	// pre means cursor position in buf
	pre int
	// post means position in buf after cursor and gap
	post int
}

func New() *GapBuffer {
	return &GapBuffer{
		buf:  make([]rune, initialsize),
		pre:  0,
		post: initialsize,
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
	copy(gb.buf[gb.post+atleast:n], gb.buf[gb.post:n])
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

func (gb *GapBuffer) Get() []rune {
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
