package editor

// XXX We have lots of buffer-checking and active-buffer-selection
//     repetition in the handler functions below. Perhaps there is a
//     way to make those prettier.

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
	"github.com/susji/ked/viewport"
)

type editorBuffer struct {
	b           *buffer.Buffer
	v           *viewport.Viewport
	lineno, col int
	linesinview int
}

type Editor struct {
	s         tcell.Screen
	buffers   []editorBuffer
	activebuf int
}

func newEditorBuffer(b *buffer.Buffer) *editorBuffer {
	return &editorBuffer{
		b:      b,
		v:      viewport.New(b),
		lineno: 0,
		col:    0,
	}
}

func New() *Editor {
	return &Editor{}
}

func (e *Editor) NewBufferFromFile(f *os.File) error {
	return e.NewBuffer(f.Name(), f)
}

func (e *Editor) NewBuffer(filepath string, r io.Reader) error {
	buf, err := buffer.NewFromReader(filepath, r)
	if err != nil {
		return err
	}
	e.buffers = append(e.buffers, *newEditorBuffer(buf))
	return nil
}

func (e *Editor) getactivebuf() *editorBuffer {
	if e.activebuf > len(e.buffers)-1 {
		panic("activebuf too large")
	}
	return &e.buffers[e.activebuf]
}

func (e *Editor) initscreen() error {
	var err error
	e.s, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := e.s.Init(); err != nil {
		return err
	}
	return nil
}

func (e *Editor) logbuffer() {
	eb := e.getactivebuf()
	for lineno := 0; lineno < eb.b.Lines(); lineno++ {
		log.Printf("[%d] %s\n", lineno, string(eb.b.GetLine(lineno).Get()))
	}
}

func (e *Editor) drawactivebuf() {
	if len(e.buffers) == 0 {
		return
	}

	eb := e.getactivebuf()
	w, h := e.s.Size()
	rend := eb.v.Render(w, h-1, eb.lineno, eb.col)
	col := 0
	lineno := 0
	for h > 0 && rend.Scan() {
		rl := rend.Line()
		for i, r := range rl.Content {
			e.s.SetContent(col+i, lineno, r, nil, tcell.StyleDefault)
		}
		lineno++
		if lineno == h-1 {
			break
		}
	}
	vx, vy := rend.Cursor()
	e.s.ShowCursor(vx, vy)
}

func (e *Editor) insertrune(r rune) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno)
	line.SetCursor(eb.col)
	line.Insert([]rune{r})
	eb.col++
}

func (e *Editor) insertlinefeed() {
	if len(e.buffers) == 0 {
		return
	}

	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno).Get()
	oldline := line[:eb.col]
	newline := line[eb.col:]

	//log.Printf("[insertlinefeed] lineno=%d/%d  oldline=%q  newline=%q\n",
	//	eb.lineno, eb.b.Lines(), oldline, newline)

	eb.b.GetLine(eb.lineno).Clear().Insert(oldline)
	eb.b.NewLine(eb.lineno + 1).Insert(newline)

	eb.lineno++
	eb.col = 0
}

func (e *Editor) backspace() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno)
	linerunes := line.Get()
	if eb.col == 0 && eb.lineno > 0 {
		eb.b.DeleteLine(eb.lineno)
		if eb.lineno > 0 {
			lineup := eb.b.GetLine(eb.lineno - 1)
			lineuprunes := lineup.Get()
			lineup.SetCursor(len(lineuprunes))
			lineup.Insert(linerunes[eb.col:])
			eb.lineno--
			eb.col = len(lineuprunes)
		}
		return
	} else if eb.col == 0 {
		return
	}
	line.SetCursor(eb.col)
	line.Delete()
	eb.col--
}

func (e *Editor) movevertical(up bool) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if up {
		if eb.lineno == 0 {
			return
		}
		eb.lineno--
	} else {
		if eb.lineno >= eb.b.Lines()-1 {
			return
		}
		eb.lineno++
	}
	line := eb.b.GetLine(eb.lineno).Get()
	if eb.col >= len(line) {
		eb.col = len(line)
	}
}

func (e *Editor) movepage(up bool) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if up {
		eb.lineno = eb.v.PageUp()

	} else {
		eb.lineno = eb.v.PageDown()
	}
}

func (e *Editor) moveleft() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if eb.col > 0 {
		eb.col--
	}
}

func (e *Editor) moveright() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno).Get()
	if eb.col < len(line) {
		eb.col++
	}
}

func (e *Editor) moveline(start bool) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if start {
		eb.col = 0
		return
	}
	line := eb.b.GetLine(eb.lineno).Get()
	eb.col = len(line)
}

func (e *Editor) savebuffer() {
	if len(e.buffers) == 0 {
		return
	}
	// XXX Ask user for a filename just to be sure
	eb := e.getactivebuf()
	log.Printf("[savebuffer] %q\n", eb.b.Filepath())
	if err := eb.b.Save(); err != nil {
		log.Println("[savebuffer] failed: ", err)
		// XXX Report error to UI somehow
	}
}

func (e *Editor) drawstatusline() {
	w, h := e.s.Size()
	fn := "{no file}"
	lineno := 0
	col := 0
	if len(e.buffers) > e.activebuf {
		eb := e.getactivebuf()
		f := eb.b.Filepath()
		if len(f) > 0 {
			fn = f
		}
		lineno = eb.lineno + 1
		col = eb.col + 1
	}
	line := []rune(
		fmt.Sprintf(
			"[%03d] %3d, %2d:  %s", e.activebuf, lineno, col, fn))
	for i, r := range line {
		e.s.SetContent(i, h-1, r, nil, tcell.StyleDefault)
		if i > w {
			break
		}
	}
}

func (e *Editor) Run() error {
	if err := e.initscreen(); err != nil {
		return err
	}
	e.drawactivebuf()
	e.s.Show()
main:
	for {
		ev := e.s.PollEvent()
		//log.Printf("[Run] event: %+v\n", ev)
		redraw := false
		sync := false
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			log.Printf("[resize] w=%d  h=%d\n", w, h)
			redraw = true
			sync = true
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[quit]")
				e.s.Fini()
				break main
			case ev.Key() == tcell.KeyRune:
				e.insertrune(ev.Rune())
				redraw = true
			case ev.Key() == tcell.KeyEnter:
				e.insertlinefeed()
				redraw = true
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				e.backspace()
				redraw = true
			case ev.Key() == tcell.KeyUp:
				e.movevertical(true)
				redraw = true
			case ev.Key() == tcell.KeyDown:
				e.movevertical(false)
				redraw = true
			case ev.Key() == tcell.KeyLeft:
				e.moveleft()
				redraw = true
			case ev.Key() == tcell.KeyRight:
				e.moveright()
				redraw = true
			case ev.Key() == tcell.KeyCtrlA:
				e.moveline(true)
				redraw = true
			case ev.Key() == tcell.KeyCtrlE:
				e.moveline(false)
				redraw = true
			case ev.Key() == tcell.KeyCtrlS:
				e.savebuffer()
				redraw = false
			case ev.Key() == tcell.KeyPgUp:
				e.movepage(true)
				redraw = true
			case ev.Key() == tcell.KeyPgDn:
				e.movepage(false)
				redraw = true
			case ev.Key() == tcell.KeyTab:
				e.insertrune('\t')
				redraw = true
			}
		}

		if redraw {
			e.s.Clear()
			e.drawactivebuf()
			e.drawstatusline()
			e.s.Show()
		}
		if sync {
			e.s.Sync()
		}
	}
	return nil
}
