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
	// y0 defines the buffer line located uppermost at the
	// moment. Note that these are *buffer* lines, not *drawn*
	// lines.
	y0 int
	// scrollup and scrolldown define the buffer lines to jump
	// into if viewport is scrolled up or down. Our viewport
	// calculation logic tries to do this such that we move in
	// jumps of screen-height/2.
	scrollup, scrolldown int
	// limitdown specifies the buffer line, which marks a
	// scroll-down. Note that this, as is y0, is in *buffer*
	// lines, not *drawn* lines.
	limitdown int
	// pageup and pagedown define the buffer lines to jump into if
	// viewport is scrolled one page up or down. As above, these
	// are also in BUFFER lines, not viewport (drawn) lines.
	pageup, pagedown int
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
}

func (h *historySumStack) Push(val int) {
	h.memory = append(h.memory, val)
}

func (h *historySumStack) CountBackwards(wantsum int) (count int, err error) {
	sum := 0
	err = errors.New("cannot find enough elements")
	for i := len(h.memory) - 1; i >= 0; i-- {
		sum += h.memory[i]
		count++
		if sum >= wantsum {
			err = nil
			break
		}
	}
	return
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
		float64(v.y0+h*2+1)))

	//
	// A rough sketch of the viewport and scrolling algorithm:
	//
	// The larger outer rectangle is the buffer we are currently
	// rendering. We do soft-wrapping here, so each line longer
	// than the viewport width is split into multiple logical,
	// viewport lines. This is where the trouble arises from: We
	// cannot jump an amount of BUFFER lines because the amount of
	// DRAWN lines may be wildly different. Thus we have to figure
	// out how many drawn lines each relevant buffer line
	// represents so we can do scroll nicely.
	//
	// Variables:
	//
	//   - `w` is viewport width in runes
	//   - `h` is viewport height in runes
	//   - `v.y0  is the current buffer line where the viewport's
	//     drawing begins from
	//   - `v.scrollup` is the buffer's linecount where we will
	//     jump if user scrolls above 'v.y0'
	//   - `v.scrolldown` is the buffer's linecount where we will
	//     jump if user scrolls below `v.limitdown`
	//   - `v.limitdown` represents the buffer line, where the
	//     viewport ends; if cursor crosses this, we jump to
	//     `v.scrolldown`
	//   - `v.pageup` and `v.pagedown` represent the buffer lines
	//      to jump to` when user wants to scroll up or down
	//      full pages
	//
	// Note: In the above explanations, "jump to" means "to set
	// the value of `v.y0`".
	//
	//                          w
	//       .-------------------------------------.
	//       |                                     |
	//       |                                     |
	//       | r     <runes before viewport>       |-v.pageup
	//       | r                                   |
	//       | r                                   |-v.scrollup
	//       | r                                   |
	//    |  +-r-----------------------------------+-v.y0
	//    |  | r                                   |
	//  h |  | r                                   |-v.scrolldown
	//    |  | r                                   |
	//    |  +-r-----------------------------------+-v.limitdown
	//       | r                                   |
	//       | r                                   |
	//       | r      <runes after viewport>       |
	//       | r                                   |-v.pagedown
	//       |                                     |
	//       '-------------------------------------'
	//
	//  `r` stands for lines which are rendered by the viewport
	//  algorithm while it's searching for the variables.
	//
	// Note: We probably have some edgecases here, which I have
	// not accounted for. I'm guessing really long lines may cause
	// wild viewport and scrolling behavior.
	//
	inview := false
	linesdrawnpreview := 0
	linesdrawninview := 0
	viewed := false
	hss := &historySumStack{}
	postviewdrawn := 0
	postviewbufs := 0
	downscrollfound := false
	// Some sane defaults to down-scroll limits. These will be
	// used if the state machine below does manage to find actual
	// viewport-crossings, ie. the buffer is not yet long enough.
	v.limitdown = v.y0 + h - 1
	v.pagedown = v.buffer.Lines() - 1
	v.scrolldown = v.y0 + h/2 - 1
scanline:
	for ; linenobuf < linenobufend; linenobuf++ {
		line := v.buffer.GetLine(linenobuf).Get()
		//log.Printf("[Render=%d] line=%q\n", linenobuf, string(line))
		newlines, _cx, _cy := v.doRenderWrapped(
			w, cursorlineno, cursorcol,
			linenobuf, linenodrawn, line)

		if linenobuf >= v.y0 && !inview && !viewed {
			// Arrived into visible viewport.
			// log.Printf("[Render] into view, linenobuf=%d\n", linenobuf)
			inview = true
			if backlines, err := hss.CountBackwards(h / 2); err == nil {
				v.scrollup = v.y0 - backlines
			} else {
				v.scrollup = 0
			}
			if backlines, err := hss.CountBackwards(h); err == nil {
				v.pageup = v.y0 - backlines
			} else {
				v.pageup = v.scrollup
			}
		} else if linesdrawninview >= h && inview && !viewed {
			// Crossed below visible viewport.
			// log.Printf("[Render] out of view, linenobuf=%d\n", linenobuf)
			inview = false
			viewed = true
			v.limitdown = linenobuf - 1
			v.pagedown = linenobuf
		} else if !inview && !viewed {
			// We keep a memo of how many preceding buffer
			// lines we need to scroll half a page
			// upwards. This matches when we are still
			// upwards from our viewport.
			hss.Push(len(newlines))
			linesdrawnpreview += len(newlines)
		} else if !inview && viewed {
			// Similar to the case of counting
			// logical/wrapped lines before entering the
			// viewport, we have to handle the
			// post-viewport thing, too. So we count how
			// many buffer lines we need to jump to get a
			// decent line to scroll the viewport down.
			// We also calculate the down-scrolling limit
			// here.
			postviewdrawn += len(newlines)
			postviewbufs++
			if postviewdrawn >= h/2 && !downscrollfound {
				v.scrolldown = v.y0 + postviewbufs
				downscrollfound = true
				break scanline
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
	log.Printf("[Render] viewed=%t  downscrollfound=%t\n",
		viewed, downscrollfound)
	log.Printf("[......] scrollup=%d  scrolldown=%d  limitdown=%d  pageup=%d  pagedown=%d\n",
		v.scrollup, v.scrolldown, v.limitdown, v.pageup, v.pagedown)
	return &Rendering{
		buf:     renderlines,
		cursorx: cx,
		cursory: cy,
	}
}

func (v *Viewport) PageUp() int {
	v.y0 = v.pageup
	return v.y0
}

func (v *Viewport) PageDown() int {
	v.y0 = v.pagedown
	return v.y0
}
