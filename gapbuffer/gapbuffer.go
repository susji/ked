package gapbuffer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
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

func NewFrom(runes []rune) *GapBuffer {
	gb := New(len(runes) + DEFAULTSZ)
	gb.SetCursor(0)
	gb.Insert(runes)
	return gb
}

func hexdump(what []rune) {
	b := &bytes.Buffer{}
	d := hex.Dumper(b)
	d.Write([]byte(string(what)))
	log.Println(b.String())
}

func debug(f string, va ...interface{}) {
	//log.Printf(f+"\n", va...)
}

func (gb *GapBuffer) Cursor() int {
	return gb.pre
}

func (gb *GapBuffer) cursorprev() {
	// |abcdefghijklmnopq/    /rstu|
	// |abcdefghijklmnop/    /qrstu|
	if gb.pre == 0 {
		panic("cursorprev: pre == 0")
	}
	gb.pre--
	gb.post--
	gb.buf[gb.post] = gb.buf[gb.pre]
}

func (gb *GapBuffer) cursornext() {
	// |abcdefghijklmnop/    /qrstu|
	// |abcdefghijklmnopq/    /rstu|
	if gb.post+1 > len(gb.buf) {
		panic("cursornext: post + 1 > len(buf)")
	}
	gb.buf[gb.pre] = gb.buf[gb.post]
	gb.pre++
	gb.post++
}

func (gb *GapBuffer) SetCursor(cursor int) *GapBuffer {
	debug("(SetCursor before) pre=%d  post=%d  {%d <- %d}", gb.pre, gb.post, cursor, gb.pre)
	if cursor > gb.Length()+1 {
		panic("cursor > gb.Length")
	}
	newpre := cursor
	var moves int
	var f func()
	if newpre < gb.pre {
		f = gb.cursorprev
		moves = gb.pre - newpre
	} else {
		f = gb.cursornext
		moves = newpre - gb.pre
	}

	for i := 0; i < moves; i++ {
		f()
	}
	debug("(SetCursor afterwards) pre=%d  post=%d", gb.pre, gb.post)
	//hexdump(gb.buf)
	return gb
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

func (gb *GapBuffer) Insert(what []rune) *GapBuffer {
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
	//hexdump(gb.buf)
	return gb
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

func (gb *GapBuffer) Clear() *GapBuffer {
	gb.buf = make([]rune, DEFAULTSZ)
	gb.pre = 0
	gb.post = DEFAULTSZ
	return gb
}

func (gb *GapBuffer) String() string {
	return fmt.Sprintf("GapBuffer(cursor=%d, contents=%q)", gb.pre, string(gb.Get()))
}
