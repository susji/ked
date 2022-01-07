package viewport

import (
	"errors"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
	"github.com/susji/ked/config"
	"github.com/susji/ked/highlighting"
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
	// paged maintains state so we don't trigger viewport
	// translation due to page up & down
	paged bool
}

type RenderLine struct {
	Content                 []rune
	LineLogical, LineBuffer int
	styles                  []tcell.Style
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

type renderedLine struct {
	content []rune
	styles  []tcell.Style
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

func (rl *RenderLine) GetStyle(col int) tcell.Style {
	if len(rl.styles) <= col {
		return config.STYLE_DEFAULT
	}
	return rl.styles[col]
}

func New(buffer *buffer.Buffer) *Viewport {
	return &Viewport{
		buffer: buffer,
	}
}

type RenderFunc func(lineno, col int, line []rune)
type CursorFunc func(lineno, col int)

func getpadding(howmuch int) []rune {
	ret := make([]rune, howmuch)
	for i := 0; i < howmuch; i++ {
		ret[i] = rune(' ')
	}
	return ret
}

func tabexpand(
	lineno int, what []rune, tabsz int, hilite *highlighting.Highlighting) (
	[]rune, []int, []tcell.Style) {

	exp := []rune("                                        ")
	new := make([]rune, 0, len(what))
	tabbedlen := make([]int, 0, len(what)+1)
	tabbedlen = append(tabbedlen, 0)
	styles := make([]tcell.Style, 0, len(what)+1)
	ntabs := 0
	for col, r := range what {
		st := hilite.Get(lineno, col)
		if r != '\t' {
			new = append(new, r)
			styles = append(styles, st)
		} else {
			new = append(new, exp[:tabsz]...)
			ntabs++
			for i := 0; i < tabsz; i++ {
				styles = append(styles, st)
			}
		}
		tabbedlen = append(tabbedlen, ntabs*(tabsz-1))
	}
	return new, tabbedlen, styles
}

func (v *Viewport) doRenderWrapped(
	w, cursorlineno, cursorcol, linenobuf, linenodrawn int, line []rune,
	hilite *highlighting.Highlighting) (
	[]renderedLine, int, int) {

	ret := []renderedLine{}
	line, tabbedlen, styles := tabexpand(linenobuf, line, v.buffer.TabSize, hilite)
	nlinefrag := int(math.Ceil(float64(len(line)) / float64(w)))

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
		stylefrag := styles[start:end]

		if endraw > end {
			// Add some padding to the last fragment to
			// have cleaner render. This could be
			// optimized faster by using, for example, a
			// static padding buffer.
			drawfrag = append(drawfrag, getpadding(endraw-end)...)
		}
		ret = append(ret, renderedLine{
			content: drawfrag,
			styles:  stylefrag,
		})

		if linenobuf == cursorlineno &&
			(cursorcol+tabbedlen[cursorcol]) >= start &&
			(cursorcol+tabbedlen[cursorcol]) <= end {
			cx = cursorcol - start + tabbedlen[cursorcol]
			cy = linenodrawn + i
		}
	}
	// Zero fragments means one line still.
	if nlinefrag == 0 {
		ret = append(ret, renderedLine{
			content: getpadding(w),
			styles:  nil,
		})
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

func (v *Viewport) Render(
	w, h, cursorlineno, cursorcol int,
	hilite *highlighting.Highlighting) *Rendering {
	if v.paged {
		v.paged = false
	} else {
		v.checktranslation(cursorlineno)
	}

	//log.Printf("[Render] w=%d  h=%d  y0=%d  cy=%d\n", w, h, v.y0, cursorlineno)
	linenodrawn := 0
	renderlines := []*RenderLine{}
	cx := 0
	cy := 0
	n := int(math.Max(
		0,
		float64(v.y0-h)))
	lastbufline := int(math.Min(
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
	const (
		VIEWPORT_BEFORE = iota
		VIEWPORT_FIRST_HALF
		VIEWPORT_SECOND_HALF
		VIEWPORT_AFTER
	)
	state := VIEWPORT_BEFORE

	linesdrawnpreview := 0
	linesdrawninview := 0
	linesbufinview := 0
	hss := &historySumStack{}
	for ; n < lastbufline && state != VIEWPORT_AFTER; n++ {
		line := v.buffer.GetLine(n)
		//log.Printf("[Render=%d] line=%q\n", linenobuf, string(line))
		renderedlines, _cx, _cy := v.doRenderWrapped(
			w, cursorlineno, cursorcol, n, linenodrawn, line, hilite)

		switch state {
		case VIEWPORT_BEFORE:
			if n >= v.y0 {
				// Arrived into visible viewport.
				// log.Printf("[Render] into view, linenobuf=%d\n", linenobuf)
				state = VIEWPORT_FIRST_HALF
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
			} else {
				// We keep a memo of how many preceding buffer
				// lines we need to scroll half a page
				// upwards. This matches when we are still
				// upwards from our viewport.
				hss.Push(len(renderedlines))
				linesdrawnpreview += len(renderedlines)
			}
		case VIEWPORT_FIRST_HALF:
			// Similar to the case of counting
			// logical/wrapped lines before entering the
			// viewport, we have to handle the
			// within-viewport thing, too. So we count how
			// many buffer lines we need to jump to get a
			// decent line to scroll the viewport down.
			v.scrolldown = v.y0 + linesbufinview
			if linesdrawninview >= h/2 {
				state = VIEWPORT_SECOND_HALF
			}
		case VIEWPORT_SECOND_HALF:
			if linesdrawninview >= h {
				// Crossed below visible viewport.
				// log.Printf("[Render] out of view, linenobuf=%d\n", linenobuf)
				state = VIEWPORT_AFTER
				v.limitdown = n - 1
				v.pagedown = n
			}
		}
		//
		// Only lines drawn within the current viewport are
		// sent back. We also have to do bookkeeping here to
		// figure out what is the actual buffer line that's
		// drawn last.
		//
		if state == VIEWPORT_FIRST_HALF || state == VIEWPORT_SECOND_HALF {
			// The scanning window thing we have here
			// hopefully gets a peek at the stuff after
			// our viewport. However, if we are operating
			// on a short buffer, or more generally at the
			// end of *any* buffer, we have to propose some
			// known-good values, because scanning
			// something after the viewport is
			// impossible. Thus we calculate some decent
			// values for `v.scrolldown` and `v.limitdown`.
			v.limitdown = n
			for ri, renderedline := range renderedlines {
				rl := &RenderLine{
					Content:     renderedline.content,
					styles:      renderedline.styles,
					LineLogical: linenodrawn + ri,
					LineBuffer:  n,
				}
				renderlines = append(renderlines, rl)
			}
			linesbufinview++
			linesdrawninview += len(renderedlines)
			if _cx != -1 && _cy != -1 {
				cx = _cx
				cy = _cy - linesdrawnpreview
			}
		}
		linenodrawn += len(renderedlines)
	}
	// It may be we have only a partial viewport to render. In
	// that case, we do not have scanned values for down-limits
	// yet. This means that we assume that the rest of the buffer
	// is filled with empty lines and deduce the correct limits.
	if state != VIEWPORT_AFTER {
		missinglines := h - linesdrawninview
		//log.Printf("[------] linesdrawinview=%d  missinglines=%d\n",
		//	linesdrawninview, missinglines)
		v.limitdown = v.y0 + linesbufinview + missinglines - 1
		v.pagedown = v.y0 + linesbufinview - 1

		// Really long lines can cause this to go negative.
		if v.limitdown < 0 {
			v.limitdown = 0
		}
	}
	//log.Printf("[Render] inview=%t  viewed=%t  downscrollfound=%t\n",
	//	inview, viewed, downscrollfound)
	//log.Printf("[......] scrollup=%d  scrolldown=%d  limitdown=%d  pageup=%d  pagedown=%d\n",
	//	v.scrollup, v.scrolldown, v.limitdown, v.pageup, v.pagedown)

	return &Rendering{
		buf:     renderlines,
		cursorx: cx,
		cursory: cy,
	}
}

func (v *Viewport) Start() int {
	return v.y0
}

func (v *Viewport) PageUp() int {
	v.paged = true
	v.y0 = v.pageup
	return v.y0
}

func (v *Viewport) PageDown() int {
	v.paged = true
	v.y0 = v.pagedown
	return v.y0
}

func (v *Viewport) SetTeleported(y int) {
	v.paged = y < v.y0 || y > v.limitdown
	if v.paged {
		v.y0 = y
	}
}
