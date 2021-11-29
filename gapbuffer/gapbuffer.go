package gapbuffer

const initialsize = 4096
const gapincrement = 32

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
		post: initialsize - gapincrement,
	}
}

func (gb *GapBuffer) Cursor() int {
	return gb.pre
}

func (gb *GapBuffer) Insert(what []rune) {
	//
	// Inserting into a GapBuffer looks like this:
	//
	//
	// /-----------------------/ gapgapgap /----------------|
	//                        pre
	//
}

func (gb *GapBuffer) Get(pos, maxlen int) ([]rune, int) {
	gaplen := gb.post - gb.pre
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

	if pos+maxlen < gb.pre {
		// #1
		ret := make([]rune, maxlen)
		copy(ret, gb.buf[pos:pos+maxlen])
		return ret, maxlen
	} else if pos >= gb.pre {
		// #2
		n := min(maxlen, len(gb.buf)-gaplen-pos)
		ret := make([]rune, n)
		return ret, n

	} else {
		// #3
	}

	ret := make([]rune, gb.pre+gb.post)
	return ret, 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
