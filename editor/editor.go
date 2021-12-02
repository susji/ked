package editor

// XXX We have lots of buffer-checking and active-buffer-selection
//     repetition in the handler functions below. Perhaps there is a
//     way to make those prettier.

import (
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
	buf, err := buffer.NewFromFile(f)
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
	rend := eb.v.Render(w, h, eb.lineno, eb.col)

	col := 0
	lineno := 0
	for rend.Scan() {
		for i, r := range rend.Line() {
			e.s.SetContent(col+i, lineno, r, nil, tcell.StyleDefault)
		}
		lineno++
		if lineno == h {
			break
		}
	}

	/*
		cf := func(lineno, col int) {
			log.Printf("[cf] lineno=%d  col=%d\n", lineno, col)
			e.s.ShowCursor(col, lineno)
		}
	*/
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

func (e *Editor) moveVertical(up bool) {
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
		if eb.lineno == eb.b.Lines() {
			return
		}
		eb.lineno++
	}
	line := eb.b.GetLine(eb.lineno).Get()
	if eb.col >= len(line) {
		eb.col = len(line)
	}
}

func (e *Editor) moveLeft() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if eb.col > 0 {
		eb.col--
	}
}

func (e *Editor) moveRight() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno).Get()
	if eb.col < len(line) {
		eb.col++
	}
}

func (e *Editor) moveLine(start bool) {
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

func (e *Editor) Run() error {
	if err := e.initscreen(); err != nil {
		return err
	}
	e.drawactivebuf()
	e.s.Show()
main:
	for {
		ev := e.s.PollEvent()
		log.Printf("[Run] event: %+v\n", ev)
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
				e.moveVertical(true)
				redraw = true
			case ev.Key() == tcell.KeyDown:
				e.moveVertical(false)
				redraw = true
			case ev.Key() == tcell.KeyLeft:
				e.moveLeft()
				redraw = true
			case ev.Key() == tcell.KeyRight:
				e.moveRight()
				redraw = true
			case ev.Key() == tcell.KeyCtrlA:
				e.moveLine(true)
				redraw = true
			case ev.Key() == tcell.KeyCtrlE:
				e.moveLine(false)
				redraw = true
			}
		}

		if redraw {
			e.s.Clear()
			e.drawactivebuf()
			e.s.Show()
		}
		if sync {
			e.s.Sync()
		}
	}
	return nil
}
