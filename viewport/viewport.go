package viewport

import (
	"log"
	"math"

	"github.com/susji/ked/buffer"
)

type Viewport struct {
	buffer *buffer.Buffer
	// x and y define the upper-left coordinates of the viewport
	x, y int
	wrap bool
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

func getPadding(howmuch int) []rune {
	ret := make([]rune, howmuch)
	for i := 0; i < howmuch; i++ {
		ret[i] = rune(' ')
	}
	return ret
}

func (v *Viewport) doRenderWrapped(w, h, lineno int, line []rune, rf RenderFunc) int {
	nlinefrag := int(math.Ceil(float64(len(line)) / float64(w)))
	log.Printf("[doRenderWrapped]: w=%d  h=%d  lineno=%d  lenline=%d   linefrags=%d\n",
		w, h, lineno, len(line), nlinefrag)

	for i := 0; i < nlinefrag; i++ {
		start := i * w
		endraw := (i + 1) * w
		end := int(math.Min(float64(endraw), float64(len(line))))
		log.Printf("[doRenderWrapped] line[%d:%d]\n", start, end)
		drawfrag := line[start:end]
		if endraw > end {
			// Add some padding to the last fragment to
			// have cleaner render. This could be
			// optimized faster by using, for example, a
			// static padding buffer.
			drawfrag = append(drawfrag, getPadding(endraw-end)...)
		}
		rf(lineno+i, 0, drawfrag)
	}
	// Zero fragments means one line still.
	if nlinefrag == 0 {
		rf(lineno, 0, getPadding(w))
		return 1
	}
	return nlinefrag
}

func (v *Viewport) Render(w, h, cx, cy int, rf RenderFunc) {
	log.Printf("[Render] w=%d  h=%d  c=(%d, %d)\n", w, h, cx, cy)
	linenodraw := v.y
	linenobuf := v.y
	for linenobuf < v.buffer.Lines() && linenodraw < h {
		line := v.buffer.GetLine(linenobuf).Get()
		log.Printf("[Render] line=%q\n", string(line))
		linenobuf++

		if v.wrap {
			linesdrawn := v.doRenderWrapped(w, h, linenodraw, line, rf)
			linenodraw += linesdrawn
		} else {
			panic("NOTIMPLEMENTED")
		}
	}
}

func (v *Viewport) SetWrapping(wrapping bool) {
	v.wrap = wrapping
}
