package viewport

import (
	"log"

	"github.com/susji/ked/buffer"
)

type Viewport struct {
	buffer *buffer.Buffer
	// x and y define the upper-left coordinates of the viewport
	x, y int
}

func New(buffer *buffer.Buffer) *Viewport {
	return &Viewport{
		buffer: buffer,
		x:      0,
		y:      0,
	}
}

type RenderFunc func(lineno, col int, line []rune)

func (v *Viewport) Render(w, h, cx, cy int, rf RenderFunc) {
	log.Printf("[Render] w=%d  h=%d  c=(%d, %d)\n", w, h, cx, cy)
	lines := v.buffer.Lines()
	for y := 0; y < h; y++ {
		lineno := y + v.y
		if lineno >= len(lines) {
			break
		}
		line := lines[lineno].Get()
		rf(lineno, 0, line)
	}
}
