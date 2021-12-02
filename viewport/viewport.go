package viewport

import (
	"errors"
	"log"
	"math"

	"github.com/susji/ked/buffer"
)

type Viewport struct {
	buffer *buffer.Buffer
	wrap   bool
	// y0 defines the buffer line located uppermost at the moment
	y0 int
	// scrollup and scrolldown define the buffer lines to jump
	// into if viewport is scrolled up or down.
	scrollup, scrolldown int
	limitdown            int
}

type RenderLine struct {
	Content                 []rune
	LineLogical, LineBuffer int
}

// Rendering essentially represents a double buffer in the graphics
// programming sense.
type Rendering struct {
	buf              []*RenderLine
	cur              *RenderLine
	done             bool
	scanned          int
	lineno           int
	cursorx, cursory int
}

func (r *Rendering) Cursor() (int, int) {
	return r.cursorx, r.cursory
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

func (r *Rendering) Line() *RenderLine {
	return r.cur
}

func New(buffer *buffer.Buffer) *Viewport {
	return &Viewport{
		buffer: buffer,
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
	w, cursorlineno, cursorcol, linenobuf, linenodrawn int, line []rune) ([][]rune, int, int) {
	ret := [][]rune{}

	nlinefrag := int(math.Ceil(float64(len(line)) / float64(w)))
	//log.Printf("[doRenderWrapped]: w=%d  h=%d  linenobuf=%d  lenline=%d   linefrags=%d\n",
	//	w, h, linenobuf, len(line), nlinefrag)

	// As we're wrapping the display, long lines need to split
	// into line fragments, which are rendered on their own
	// terminal rows. Also, we need similar logic to figure out
	// our cursor position.
	cx := -1
	cy := -1
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
		if linenobuf == cursorlineno && cursorcol >= start && cursorcol <= end {
			cx = cursorcol - start
			cy = linenodrawn + i
		}
	}
	// Zero fragments means one line still.
	if nlinefrag == 0 {
		ret = append(ret, getPadding(w))
		if linenobuf == cursorlineno {
			cx = 0
			cy = linenodrawn
		}
	}
	return ret, cx, cy
}

type historySumStack struct {
	memory []int
	done   bool
}

func (h *historySumStack) Push(val int) {
	if h.done {
		panic("historySumStack: stack already exhausted")
	}
	h.memory = append(h.memory, val)
}

func (h *historySumStack) CountBackwards(wantsum int) (int, error) {
	if h.done {
		panic("historySumStack: can count only once")
	}
	sum := 0
	popped := 0
	for {
		if len(h.memory) == 0 {
			h.done = true
			return -1, errors.New("memory ran out")
		}
		sum += h.memory[len(h.memory)-1]
		popped++
		h.memory = h.memory[1:]
		if sum >= wantsum {
			break
		}
	}
	h.done = true
	return popped, nil
}

func (v *Viewport) checktranslation(cursorlineno int) {
	if cursorlineno < v.y0 {
		v.y0 = v.scrollup
	} else if cursorlineno > v.limitdown {
		v.y0 = v.scrolldown
	}
}

func (v *Viewport) Render(w, h, cursorlineno, cursorcol int) *Rendering {
	v.checktranslation(cursorlineno)

	log.Printf("[Render] w=%d  h=%d  y0=%d  cy=%d\n", w, h, v.y0, cursorlineno)
	linenodrawn := 0
	renderlines := []*RenderLine{}
	cx := 0
	cy := 0
	linenobuf := int(math.Max(
		0,
		float64(v.y0-h)))
	linenobufend := int(math.Min(
		float64(v.buffer.Lines()),
		float64(v.y0+h*2)))

	//
	// Below you'll see a loop where we render buffer lines
	// beginning from somewhere before the current cursor line all
	// the way beyond a few pages after our cursor. The point of
	// this is to find suitable scroll-positions: If we render
	// with soft-wrapping, a single "buffer line" may produce an
	// arbitrarily large amount of "logical" or "drawing
	// lines". This means that we have to calculate our
	// scroll-lines based on the "logical" lines. Hence all this
	// wrestling.
	//
	inview := false
	linesdrawnpreview := 0
	linesdrawninview := 0
	viewed := false
	hss := &historySumStack{}
	postviewdrawn := 0
	postviewbufs := 0
	downscrolldone := false
	// Some sane defaults to down-scroll limits. These will be
	// used if the state machine below does manage to find actual
	// viewport-crossings, ie. the buffer is not yet long enough.
	v.limitdown = v.y0 + h - 3
	v.scrolldown = v.y0 + h/2 - 1
	for ; linenobuf < linenobufend; linenobuf++ {
		line := v.buffer.GetLine(linenobuf).Get()
		//log.Printf("[Render=%d] line=%q\n", linenobuf, string(line))
		newlines, _cx, _cy := v.doRenderWrapped(
			w, cursorlineno, cursorcol,
			linenobuf, linenodrawn, line)

		if linenobuf >= v.y0 && !inview && !viewed {
			// Arrived into visible viewport.
			log.Printf("[Render] into view, linenobuf=%d\n",
				linenobuf)
			inview = true
			if backlines, err := hss.CountBackwards(h / 2); err == nil {
				v.scrollup = v.y0 - backlines
			} else {
				v.scrollup = 0
			}
		} else if linesdrawninview >= h && inview && !viewed {
			// Crossed below visible viewport.
			log.Printf("[Render] out of view, linenobuf=%d\n",
				linenobuf)
			inview = false
			viewed = true
		} else if !inview && !viewed {
			// We keep a memo of how many preceding buffer
			// lines we need to scroll half a page
			// upwards. This matches when we are still
			// upwards from our viewport.
			hss.Push(len(newlines))
			linesdrawnpreview += len(newlines)
		} else if !inview && viewed && !downscrolldone {
			// Similar to the case of counting
			// logical/wrapped lines before entering the
			// viewport, we have to handle the
			// post-viewport thing, too. So we count how
			// many buffer lines we need to jump to get a
			// decent line to scroll the viewport down.
			postviewdrawn += len(newlines)
			postviewbufs++
			if postviewdrawn >= h/2 {
				v.scrolldown = v.y0 + postviewbufs
				downscrolldone = true
			}
		}
		//
		// Only lines drawn within the current viewport are
		// sent back. We also have to do bookkeeping here to
		// figure out what is the actual buffer that's drawn
		// last.
		//
		if inview {
			for ri, linecontent := range newlines {
				rl := &RenderLine{
					Content:     linecontent,
					LineLogical: linenodrawn + ri,
					LineBuffer:  linenobuf,
				}
				renderlines = append(renderlines, rl)
			}
			linesdrawninview += len(newlines)
			if _cx != -1 && _cy != -1 {
				cx = _cx
				cy = _cy - linesdrawnpreview
			}
		}
		linenodrawn += len(newlines)
	}
	log.Printf("[Render] scrollup=%d  scrolldown=%d  limitdown=%d\n",
		v.scrollup, v.scrolldown, v.limitdown)
	return &Rendering{
		buf:     renderlines,
		cursorx: cx,
		cursory: cy,
	}
}
