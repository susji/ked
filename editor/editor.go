package editor

// XXX We have lots of buffer-checking and active-buffer-selection
//     repetition in the handler functions below. Perhaps there is a
//     way to make those prettier.

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
	"github.com/susji/ked/editor/buffers"
	"github.com/susji/ked/fuzzyselect"
	"github.com/susji/ked/textentry"
)

type Editor struct {
	s         tcell.Screen
	buffers   buffers.EditorBuffers
	activebuf buffers.BufferId
	savehook  string

	prevsearch map[buffers.BufferId]string
}

func New() *Editor {
	return &Editor{
		prevsearch: map[buffers.BufferId]string{},
		buffers:    buffers.New(),
	}
}

func (e *Editor) NewBufferFromFile(f *os.File) (buffers.BufferId, error) {
	return e.NewBuffer(f.Name(), f)
}

func (e *Editor) NewBuffer(filepath string, r io.Reader) (buffers.BufferId, error) {
	buf, err := buffer.NewFromReader(filepath, r)
	if err != nil {
		return 0, err
	}
	return e.NewFromBuffer(buf)
}

func (e *Editor) NewFromBuffer(buf *buffer.Buffer) (buffers.BufferId, error) {
	bid := e.buffers.New(buf)
	e.activebuf = bid
	return bid, nil
}

func (e *Editor) SaveHook(savehook string) *Editor {
	e.savehook = savehook
	return e
}

func (e *Editor) closebuffer(bufid buffers.BufferId) {
	e.buffers.Close(bufid)
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
	eb := e.buffers.Get(e.activebuf)
	w, h := e.s.Size()
	rend := eb.Viewport.Render(w, h-1, eb.CursorLine(), eb.CursorCol())
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
	eb := e.buffers.Get(e.activebuf)
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewInsert(eb.CursorLine(), eb.CursorCol(), []rune{r})))
}

func (e *Editor) insertlinefeed() {
	eb := e.buffers.Get(e.activebuf)
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewLinefeed(eb.Cursor())))
}

func (e *Editor) backspace() {
	eb := e.buffers.Get(e.activebuf)
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewBackspace(eb.Cursor())))
}

func (e *Editor) movevertical(up bool) {
	eb := e.buffers.Get(e.activebuf)
	if up {
		if eb.CursorLine() == 0 {
			return
		}
		eb.SetCursor(eb.CursorLine()-1, eb.CursorCol())
	} else {
		if eb.CursorLine() >= eb.Buffer.Lines()-1 {
			return
		}
		eb.SetCursor(eb.CursorLine()+1, eb.CursorCol())
	}
	line := eb.Buffer.GetLine(eb.CursorLine())
	if eb.CursorCol() >= len(line) {
		eb.SetCursor(eb.CursorLine(), len(line))
	}
}

func (e *Editor) movepage(up bool) {
	eb := e.buffers.Get(e.activebuf)
	if up {
		eb.SetCursor(eb.Viewport.PageUp(), eb.CursorCol())
	} else {
		eb.SetCursor(eb.Viewport.PageDown(), eb.CursorCol())
	}
	eb.SetCursor(eb.CursorLine(), 0)
}

func (e *Editor) moveleft() {
	eb := e.buffers.Get(e.activebuf)
	if eb.CursorCol() > 0 {
		eb.SetCursor(eb.CursorLine(), eb.CursorCol()-1)
	} else if eb.CursorLine() > 0 {
		lineno := eb.CursorLine() - 1
		eb.SetCursor(lineno, len(eb.Buffer.GetLine(lineno)))
	}
}

func (e *Editor) moveright() {
	eb := e.buffers.Get(e.activebuf)
	line := eb.Buffer.GetLine(eb.CursorLine())
	if eb.CursorCol() < len(line) {
		eb.SetCursor(eb.CursorLine(), eb.CursorCol()+1)
	} else if eb.CursorLine() < eb.Buffer.Lines()-1 {
		eb.SetCursor(eb.CursorLine()+1, 0)
	}
}

func (e *Editor) moveline(start bool) {
	eb := e.buffers.Get(e.activebuf)
	if start {
		eb.SetCursor(eb.CursorLine(), 0)
		return
	}
	line := eb.Buffer.GetLine(eb.CursorLine())
	eb.SetCursor(eb.CursorLine(), len(line))
}

func (e *Editor) savebuffer() {
	eb := e.buffers.Get(e.activebuf)
	log.Printf("[savebuffer] %q\n", eb.Buffer.Filepath())

	_, h := e.s.Size()
	fp, err := textentry.
		New(eb.Buffer.Filepath(), "Filename to save: ", 512).
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
	eb.Buffer.SetFilepath(abspath)
	if err := eb.Buffer.Save(); err != nil {
		log.Println("[savebuffer] failed: ", err)
		// XXX Report error to UI somehow
	}

	if len(e.savehook) > 0 {
		sh := strings.ReplaceAll(e.savehook, "__ABSPATH__", abspath)
		log.Printf("[savebuffer, hook] %q -> %q\n", e.savehook, sh)
		e.execandreload(abspath, sh)
	}
}

func (e *Editor) execandreload(abspath, cmd string) {
	args := []string{"-c", cmd}
	c := exec.Command("/bin/sh", args...)
	out, err := c.Output()
	log.Println("[execandreread, output] ", out)
	if err != nil {
		// XXX Display error to user somehow.
		log.Printf("[execandreread, exec error] %v\n", err)
		return
	}
	f, err := os.Open(abspath)
	if err != nil {
		// XXX Display error to use somehow.
		log.Printf("[execandreread, reopen error] %v\n", err)
		return
	}
	defer f.Close()

	oldbuf := e.buffers.Get(e.activebuf)
	oldviewportstart := oldbuf.Viewport.Start()
	oldcursorline, oldcursorcol := oldbuf.CursorLine(), oldbuf.CursorCol()

	log.Println("[execandreread, reopened] ", abspath)
	newbid, err := e.NewBufferFromFile(f)
	if err != nil {
		log.Printf("[execandreread, newbuffer error] %v\n", err)
		return
	}
	e.buffers.Close(oldbuf.Id())
	newbuf := e.buffers.Get(newbid)

	//
	// Make sure old cursor snaps into new buffer.
	//
	newcursorline := oldcursorline
	newcursorcol := oldcursorcol
	if newcursorline > newbuf.Buffer.Lines()-1 {
		newcursorline = newbuf.Buffer.Lines() - 1
	}
	if newcursorcol > newbuf.Buffer.LineLength(newcursorline) {
		newcursorcol = newbuf.Buffer.LineLength(newcursorline)
	}
	if oldviewportstart > newbuf.Buffer.Lines()-1 {
		oldviewportstart = newbuf.Buffer.Lines() - 1
	}

	newbuf.SetCursor(newcursorline, newcursorcol)
	newbuf.Viewport.SetTeleported(oldviewportstart)
}

func (e *Editor) jumpline() {
	eb := e.buffers.Get(e.activebuf)
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
	if lineno > eb.Buffer.Lines() {
		lineno = eb.Buffer.Lines()
	}
	eb.SetCursor(lineno-1, 0)
}

func (e *Editor) delline() {
	eb := e.buffers.Get(e.activebuf)
	if eb.Buffer.LineLength(eb.CursorLine()) == 0 {
		eb.Update(eb.Buffer.Perform(buffer.NewDelLine(eb.CursorLine())))
		return
	}
	eb.Update(eb.Buffer.Perform(buffer.NewDelLineContent(eb.CursorLine(), eb.CursorCol())))
}

func (e *Editor) jumpword(left bool) {
	eb := e.buffers.Get(e.activebuf)
	eb.SetCursor(eb.Buffer.JumpWord(eb.CursorLine(), eb.CursorCol(), left))
}

func (e *Editor) search() {
	eb := e.buffers.Get(e.activebuf)

	if _, ok := e.prevsearch[eb.Id()]; !ok {
		e.prevsearch[eb.Id()] = ""
	}

	_, h := e.s.Size()
	nexterr := errors.New("next term")
	te := textentry.
		New(e.prevsearch[eb.Id()], "Search: ", 256).
		AddBinding(tcell.KeyCtrlS, nexterr)
	looping := true
	prevcol := eb.CursorCol()
	for looping {
		term, err := te.Ask(e.s, 0, h-1)
		switch {
		case err == nil:
			looping = false
		case errors.Is(err, nexterr):
			log.Printf("[search] got next for %q\n", string(term))
		default:
			log.Println("[search, error-ask] ", err)
			e.drawstatusmsg(fmt.Sprintf("%v", err))
			return
		}

		sterm := []rune{}
		for _, r := range term {
			sterm = append(sterm, unicode.ToLower(r))
		}
		limits := &buffer.SearchLimit{
			StartLineno: eb.CursorLine(),
			StartCol:    prevcol,
			EndLineno:   eb.Buffer.Lines() - 1,
			EndCol:      eb.Buffer.LineLength(eb.Buffer.Lines() - 1),
		}
		log.Printf("[search, limits] %#v\n", limits)
		if lineno, col := eb.Buffer.SearchRange(sterm, limits); lineno != -1 && col != -1 {
			log.Printf("[search, found] (%d, %d)\n", lineno, col)
			eb.SetCursor(lineno, col)
			eb.Viewport.SetTeleported(eb.CursorLine())
			e.prevsearch[eb.Id()] = string(term)
			e.s.Clear()
			e.drawactivebuf()
			e.s.Show()

			prevcol = eb.CursorCol() + len(term)
			linelen := eb.Buffer.LineLength(lineno)
			if prevcol >= linelen {
				prevcol = linelen - 1
			}
		}
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

func (e *Editor) listbuffers() {
	log.Println("[listbuffers]")
	for bufnum, buf := range e.buffers.All() {
		log.Printf("[%03d] %#v\n", bufnum, buf)
	}
}

func (e *Editor) drawstatusline() {
	w, h := e.s.Size()
	fn := "{no file}"
	lineno := 0
	col := 0
	eb := e.buffers.Get(e.activebuf)
	f := eb.Buffer.Filepath()
	if len(f) > 0 {
		fn = f
	}
	lineno = eb.CursorLine() + 1
	col = eb.CursorCol() + 1
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

func (e *Editor) jumpempty(up bool) {
	eb := e.buffers.Get(e.activebuf)
moveagain:
	e.movevertical(up)
	if strings.TrimSpace(string(eb.Buffer.GetLine(eb.CursorLine()))) == "" {
		return
	}
	if up && eb.CursorLine() == 0 {
		return
	}
	if !up && eb.CursorLine() >= eb.Buffer.Lines()-1 {
		return
	}
	goto moveagain
}

func (e *Editor) undo() {
	eb := e.buffers.Get(e.activebuf)
	if res := eb.Buffer.UndoModification(); res != nil {
		eb.Update(*res)
	}
}

func (e *Editor) backtab() {
	eb := e.buffers.Get(e.activebuf)
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewDetabulate(eb.CursorLine(), eb.CursorCol())))
}

func (e *Editor) changebuffer() {
	choices := []fuzzyselect.Entry{}

	for bufnum, bufentry := range e.buffers.All() {
		choices = append(choices, fuzzyselect.Entry{
			Display: []rune(bufentry.Buffer.Filepath()),
			Id:      uint32(bufnum),
		})
	}

	w, h := e.s.Size()
	sel, err := fuzzyselect.New(choices).Choose(e.s, 0, 0, w, h-2)
	if err != nil {
		// XXX Display error to user somehow.
		log.Printf("[changebuffer, fuzzy error] %v\n", err)
		return
	}
	e.activebuf = buffers.BufferId(sel.Id)
}

func (e *Editor) Run() error {
	if err := e.initscreen(); err != nil {
		return err
	}
	if e.buffers.Len() == 0 {
		e.NewFromBuffer(buffer.New(nil))
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
			case ev.Key() == tcell.KeyCtrlP:
				e.changebuffer()
			case ev.Key() == tcell.KeyCtrlL:
				e.listbuffers()
			case ev.Key() == tcell.KeyCtrlUnderscore:
				e.undo()
			case ev.Key() == tcell.KeyCtrlS:
				e.search()
			case ev.Key() == tcell.KeyCtrlK:
				e.delline()
			case ev.Key() == tcell.KeyCtrlG:
				e.jumpline()
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyUp:
				e.jumpempty(true)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyDown:
				e.jumpempty(false)
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
			case ev.Key() == tcell.KeyCtrlW:
				e.savebuffer()
			case ev.Key() == tcell.KeyPgUp:
				e.movepage(true)
			case ev.Key() == tcell.KeyPgDn:
				e.movepage(false)
			case ev.Key() == tcell.KeyTab:
				e.insertrune('\t')
			case ev.Key() == tcell.KeyBacktab:
				e.backtab()
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
