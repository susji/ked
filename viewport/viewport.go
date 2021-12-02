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

func (v *Viewport) doRenderWrapped(w, h, cursorlineno, cursorcol,
	linenobuf, linenodraw int, line []rune,
	rf RenderFunc, cf CursorFunc) int {

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
		// log.Printf("[doRenderWrapped] line[%d:%d]\n", start, end)
		drawfrag := line[start:end]
		if endraw > end {
			// Add some padding to the last fragment to
			// have cleaner render. This could be
			// optimized faster by using, for example, a
			// static padding buffer.
			drawfrag = append(drawfrag, getPadding(endraw-end)...)
		}
		rf(linenodraw+i, 0, drawfrag)
		if linenobuf == cursorlineno && cursorcol >= start && cursorcol <= end {
			cf(linenodraw+i, cursorcol-start)
		}
	}
	// Zero fragments means one line still.
	if nlinefrag == 0 {
		rf(linenodraw, 0, getPadding(w))
		cf(linenodraw, 0)
		return 1
	}
	return nlinefrag
}

func (v *Viewport) Render(w, h, cursorlineno, cursorcol int, rf RenderFunc, cf CursorFunc) {
	//log.Printf("[Render] w=%d  h=%d  c=(%d, %d)\n", w, h, cx, cy)
	linenodraw := v.y
	linenobuf := v.y
	for linenobuf < v.buffer.Lines() && linenodraw < h {
		line := v.buffer.GetLine(linenobuf).Get()
		//log.Printf("[Render] line=%q\n", string(line))
		if v.wrap {
			linesdrawn := v.doRenderWrapped(
				w, h, cursorlineno, cursorcol,
				linenobuf, linenodraw, line, rf, cf)
			linenodraw += linesdrawn
		} else {
			panic("NOTIMPLEMENTED")
		}
		linenobuf++
	}
}

func (v *Viewport) SetWrapping(wrapping bool) {
	v.wrap = wrapping
}
