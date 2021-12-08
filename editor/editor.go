package editor

// XXX We have lots of buffer-checking and active-buffer-selection
//     repetition in the handler functions below. Perhaps there is a
//     way to make those prettier.

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
	"github.com/susji/ked/textentry"
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

func (eb *editorBuffer) update(res buffer.ActionResult) {
	eb.lineno = res.Lineno
	eb.col = res.Col
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
		log.Printf("[%d] %s\n", lineno, string(eb.b.GetLine(lineno)))
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
	eb.update(eb.b.Perform(buffer.NewInsert(eb.lineno, eb.col, []rune{r})))
}

func (e *Editor) insertlinefeed() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	eb.update(eb.b.Perform(buffer.NewLinefeed(eb.lineno, eb.col)))
}

func (e *Editor) backspace() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	eb.update(eb.b.Perform(buffer.NewBackspace(eb.lineno, eb.col)))
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
	line := eb.b.GetLine(eb.lineno)
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
	eb.col = 0
}

func (e *Editor) moveleft() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if eb.col > 0 {
		eb.col--
	} else if eb.lineno > 0 {
		eb.lineno--
		eb.col = len(eb.b.GetLine(eb.lineno))
	}
}

func (e *Editor) moveright() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	line := eb.b.GetLine(eb.lineno)
	if eb.col < len(line) {
		eb.col++
	} else if eb.lineno < eb.b.Lines()-1 {
		eb.lineno++
		eb.col = 0
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
	line := eb.b.GetLine(eb.lineno)
	eb.col = len(line)
}

func (e *Editor) savebuffer() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	log.Printf("[savebuffer] %q\n", eb.b.Filepath())

	_, h := e.s.Size()
	fp, err := textentry.
		New(eb.b.Filepath(), "Filename to save: ", 512).
		Ask(e.s, 0, h-1)
	if err != nil {
		log.Println("[savebuffer, error-ask] ", err)
		e.drawstatusmsg(fmt.Sprintf("%v", err))
		return
	}
	abspath, err := filepath.Abs(string(fp))
	if err != nil {
		log.Println("[savebuffer, error-abs] ", err)
		e.drawstatusmsg(fmt.Sprintf("%v", err))
		return
	}
	log.Println("[savebuffer, abs] ", abspath)
	eb.b.SetFilepath(abspath)
	if err := eb.b.Save(); err != nil {
		log.Println("[savebuffer] failed: ", err)
		// XXX Report error to UI somehow
	}
}

func (e *Editor) jumpline() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	_, h := e.s.Size()
	linenoraw, err := textentry.
		New("", "Line to jump: ", 12).
		Ask(e.s, 0, h-1)
	if err != nil {
		log.Println("[jumpline, error-ask] ", err)
		return
	}
	lineno, err := strconv.Atoi(string(linenoraw))
	if err != nil {
		log.Println("[jumpline, error-conv] ", err)
		return
	}
	if lineno < 1 {
		log.Println("[jumpline, invalid line] ", lineno)
		return
	}
	if lineno > eb.b.Lines() {
		lineno = eb.b.Lines()
	}
	eb.lineno = lineno - 1
	eb.col = 0
	eb.v.SetTeleported(eb.lineno)
}

func (e *Editor) delline() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if eb.b.LineLength(eb.lineno) == 0 {
		eb.update(eb.b.Perform(buffer.NewDelLine(eb.lineno)))
		return
	}
	eb.update(eb.b.Perform(buffer.NewDelLineContent(eb.lineno, eb.col)))
}

func (e *Editor) jumpword(left bool) {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	eb.lineno, eb.col = eb.b.JumpWord(eb.lineno, eb.col, left)
}

func (e *Editor) search() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	_, h := e.s.Size()
	term, err := textentry.New("", "Search: ", 256).Ask(e.s, 0, h-1)
	if err != nil {
		log.Println("[search, error-ask] ", err)
		e.drawstatusmsg(fmt.Sprintf("%v", err))
		return
	}
	sterm := []rune{}
	for _, r := range term {
		sterm = append(sterm, unicode.ToLower(r))
	}
	limits := &buffer.SearchLimit{
		StartLineno: eb.lineno,
		StartCol:    eb.col,
		EndLineno:   eb.b.Lines() - 1,
		EndCol:      eb.b.LineLength(eb.b.Lines() - 1),
	}
	if lineno, col := eb.b.SearchRange(sterm, limits); lineno != -1 && col != -1 {
		eb.lineno = lineno
		eb.col = col
		eb.v.SetTeleported(eb.lineno)
	}
}

func (e *Editor) drawstatusmsg(msg string) {
	log.Println("[drawstatusmsg] ", msg)
	w, h := e.s.Size()
	for i, r := range []rune(msg) {
		e.s.SetContent(i, h-1, r, nil, tcell.StyleDefault)
		if i > w {
			break
		}
	}
	e.s.Show()
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

func (e *Editor) undo() {
	if len(e.buffers) == 0 {
		return
	}
	eb := e.getactivebuf()
	if res := eb.b.UndoModification(); res != nil {
		eb.update(*res)
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
		sync := false
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			log.Printf("[resize] w=%d  h=%d\n", w, h)
			sync = true
		case *tcell.EventKey:
			log.Printf("[EventKey] %s (mods=%X)\n", ev.Name(), ev.Modifiers())
			switch {
			case ev.Key() == tcell.KeyCtrlUnderscore:
				e.undo()
			case ev.Key() == tcell.KeyCtrlF:
				e.search()
			case ev.Key() == tcell.KeyCtrlK:
				e.delline()
			case ev.Key() == tcell.KeyCtrlG:
				e.jumpline()
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyLeft:
				e.jumpword(true)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyRight:
				e.jumpword(false)
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[quit]")
				e.s.Fini()
				break main
			case ev.Key() == tcell.KeyRune:
				e.insertrune(ev.Rune())
			case ev.Key() == tcell.KeyEnter:
				e.insertlinefeed()
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				e.backspace()
			case ev.Key() == tcell.KeyUp:
				e.movevertical(true)
			case ev.Key() == tcell.KeyDown:
				e.movevertical(false)
			case ev.Key() == tcell.KeyLeft:
				e.moveleft()
			case ev.Key() == tcell.KeyRight:
				e.moveright()
			case ev.Key() == tcell.KeyCtrlA:
				e.moveline(true)
			case ev.Key() == tcell.KeyCtrlE:
				e.moveline(false)
			case ev.Key() == tcell.KeyCtrlS:
				e.savebuffer()
			case ev.Key() == tcell.KeyPgUp:
				e.movepage(true)
			case ev.Key() == tcell.KeyPgDn:
				e.movepage(false)
			case ev.Key() == tcell.KeyTab:
				e.insertrune('\t')
			}
		}

		e.s.Clear()
		e.drawactivebuf()
		e.drawstatusline()
		e.s.Show()

		if sync {
			e.s.Sync()
		}
	}
	return nil
}
