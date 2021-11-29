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
	gaplen := gb.post - gb.pre
	//
	// Inserting into a GapBuffer looks like this:
	//
	//                           what
	//                         -======-
	//
	// /-----------------------/ gapgapgap /----------------|
	//                        pre
	//
	if gaplen <= len(what) {
		gb.gapgrow(len(what))
	}
	copy(gb.buf[gb.pre:], what)
	gb.pre += len(what)
	hexdump(gb.buf)
}

func (gb *GapBuffer) Get(pos, maxlen int) ([]rune, int) {
	gaplen := gb.post - gb.pre
	debug("(Get) pos=%d  maxlen=%d  gaplen=%d", pos, maxlen, gaplen)
	if pos >= len(gb.buf)-gaplen {
		// We consider it a misuse of the API and thus a bug
		// if we're asked for bytes beyond the end.
		panic("Get: pos > len(buf)")
	}
	//
	// Getting a slice out of our GapBuffer contents looks like
	// this:
	//
	//
	//               pos                         maxlen
	//
	//                sliceslice           sliceslice
	//                .========-           -========.
	//                |                             |
	// /--------------+--------/ gapgapgap /--------+-------|
	//               l1       l2           r1       r2
	//
	// 0                     pre         post            len(buf)
	//
	//
	// So, we need to fish out sub-buffers [l1:l2] and [r1:r2], of
	// which either may be empty.  All in all, we have three
	// different cases here:
	//
	//   #1  Request is completely on the left side of the gap
	//   #2  Request is completely on the right side of the gap
	//   #3  Request contains both sides of the gap
	//

	if pos+maxlen <= gb.pre {
		// #1
		debug("case #1")
		start := pos
		n := maxlen
		end := pos + maxlen

		debug("[%d, %d]=%d", start, end, n)

		ret := make([]rune, n)
		copy(ret, gb.buf[start:end])
		return ret, n
	} else if pos >= gb.pre {
		// #2
		debug("case #2")
		start := pos + gaplen
		n := maxlen
		end := start + n
		overreach := end - len(gb.buf)
		if overreach > 0 {
			n -= overreach
			end -= overreach
		}

		debug("[%d, %d]=%d", start, end, n)

		ret := make([]rune, n)
		copy(ret, gb.buf[start:end])
		return ret, n

	} else {
		// #3
		debug("case #3")
		start1 := pos
		n1 := gb.pre - pos
		end1 := gb.pre

		start2 := gb.post
		n2 := maxlen - n1
		end2 := gb.post + n2
		overreach := end2 - len(gb.buf)
		if overreach > 0 {
			n2 -= overreach
			end2 -= overreach
		}

		debug("[%d, %d]=%d  [%d, %d]=%d", start1, end1, n1, start2, end2, n2)

		ret := make([]rune, n1+n2)
		copy(ret, gb.buf[start1:end1])
		copy(ret[n2:], gb.buf[start2:end2])
		return ret, n1 + n2
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
