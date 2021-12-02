package viewport

import (
	"math"

	"github.com/susji/ked/buffer"
)

type Viewport struct {
	buffer *buffer.Buffer
	// x and y define the upper-left coordinates of the viewport
	x, y int
	wrap bool
	// cursorsx and cursory define the latest known
	// screen-coordinates for our cursor
	cursorx, cursory int
}

// Rendering essentially represents a double buffer in the graphics
// programming sense.
type Rendering struct {
	buf              [][]rune
	cur              []rune
	done             bool
	scanned          int
	cursorx, cursory int
}

func (r *Rendering) Scan() bool {
	//log.Printf("[Scan] done=%t  cur=%q  buf=%q\n", r.done, string(r.cur), r.buf)
	if r.done {
		return false
	}
	if len(r.buf) == 0 {
		r.done = true
		r.cur = nil
		return false
	}
	r.cur = r.buf[0]
	if len(r.buf) > 0 {
		r.buf = r.buf[1:]
	} else {
		r.cur = nil
	}
	return true
}

func (r *Rendering) Line() []rune {
	return r.cur
}

func New(buffer *buffer.Buffer) *Viewport {
	return &Viewport{
		buffer: buffer,
		x:      0,
		y:      0,
		wrap:   true,
	}
}

type RenderFunc func(lineno, col int, line []rune)
type CursorFunc func(lineno, col int)

func getPadding(howmuch int) []rune {
	ret := make([]rune, howmuch)
	for i := 0; i < howmuch; i++ {
		ret[i] = rune(' ')
	}
	return ret
}

func (v *Viewport) doRenderWrapped(
	w, cursorlineno, cursorcol, linenobuf, linenodraw int, line []rune) [][]rune {
	ret := [][]rune{}

	nlinefrag := int(math.Ceil(float64(len(line)) / float64(w)))
	//log.Printf("[doRenderWrapped]: w=%d  h=%d  linenobuf=%d  lenline=%d   linefrags=%d\n",
	//	w, h, linenobuf, len(line), nlinefrag)

	// As we're wrapping the display, long lines need to split
	// into line fragments, which are rendered on their own
	// terminal rows. Also, we need similar logic to figure out
	// our cursor position.
	for i := 0; i < nlinefrag; i++ {
		start := i * w
		endraw := (i + 1) * w
		end := int(math.Min(float64(endraw), float64(len(line))))
		drawfrag := line[start:end]
		//log.Printf("[doRenderWrapped] line[%d:%d]=%q\n",
		//	start, end, string(drawfrag))
		if endraw > end {
			// Add some padding to the last fragment to
			// have cleaner render. This could be
			// optimized faster by using, for example, a
			// static padding buffer.
			drawfrag = append(drawfrag, getPadding(endraw-end)...)
		}
		ret = append(ret, drawfrag)
		//		if linenobuf == cursorlineno && cursorcol >= start && cursorcol <= end {
		//	cf(linenodraw+i, cursorcol-start)
		//}
	}
	// Zero fragments means one line still.
	if nlinefrag == 0 {
		ret = append(ret, getPadding(w))
		//cf(linenodraw, 0)
	}
	return ret
}

func (v *Viewport) Render(w, h, cursorlineno, cursorcol int) *Rendering {
	//log.Printf("[Render] w=%d  h=%d  c=(%d, %d)\n", w, h, cx, cy)
	// XXX Temporarily render all lines.
	linenodrawn := 0
	renderlines := [][]rune{}
	for linenobuf := 0; linenobuf < v.buffer.Lines(); linenobuf++ {
		line := v.buffer.GetLine(linenobuf).Get()
		//log.Printf("[Render=%d] line=%q\n", linenobuf, string(line))
		// XXX We only do line-wrapped mode here.
		newlines := v.doRenderWrapped(
			w, cursorlineno, cursorcol,
			linenobuf, linenodrawn, line)
		linenodrawn += len(newlines)
		renderlines = append(renderlines, newlines...)
	}
	return &Rendering{buf: renderlines}
}

func (v *Viewport) SetWrapping(wrapping bool) {
	v.wrap = wrapping
}
