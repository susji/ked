package editor

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
		lineno: 1,
		col:    1,
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

func (e *Editor) drawactivebuf() {
	if len(e.buffers) == 0 {
		return
	}

	rf := func(lineno, col int, linerunes []rune) {
		for i, r := range linerunes {
			e.s.SetContent(col+i, lineno, r, nil, tcell.StyleDefault)
		}
	}

	eb := e.getactivebuf()
	w, h := e.s.Size()
	eb.v.Render(w, h, 0, 0, rf)
	e.s.Sync()
}

func (e *Editor) insertrune(r rune) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.Lines()[eb.lineno-1]
	col := eb.col - 1
	line.SetCursor(col)
	line.Insert([]rune{r})
	eb.col++
}

func (e *Editor) Run() error {

	if err := e.initscreen(); err != nil {
		return err
	}
	e.drawactivebuf()
main:
	for {
		e.s.Show()
		ev := e.s.PollEvent()
		log.Printf("event: %+v\n", ev)
		redraw := false
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			log.Printf("[resize] w=%d  h=%d\n", w, h)
			redraw = true
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[quit]")
				e.s.Fini()
				break main
			case ev.Key() == tcell.KeyRune:
				e.insertrune(ev.Rune())
				redraw = true
			}

		}
		if redraw {
			e.drawactivebuf()
		}

	}
	return nil
}