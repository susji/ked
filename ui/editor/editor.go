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

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
	"github.com/susji/ked/config"
	"github.com/susji/ked/highlighting"
	"github.com/susji/ked/library"
	"github.com/susji/ked/ui/dialog"
	"github.com/susji/ked/ui/editor/buffers"
	"github.com/susji/ked/ui/fuzzyselect"
	"github.com/susji/ked/ui/textentry"
	"github.com/susji/ked/util"
)

type Editor struct {
	s         tcell.Screen
	buffers   buffers.EditorBuffers
	activebuf buffers.BufferId

	prevopendir   string
	prevsearch    map[buffers.BufferId]string
	bufpopularity map[buffers.BufferId]uint64
	modified      map[buffers.BufferId]bool
}

func New() *Editor {
	return NewWithScreen(nil)
}

func NewWithScreen(s tcell.Screen) *Editor {
	return &Editor{
		prevsearch:    map[buffers.BufferId]string{},
		bufpopularity: map[buffers.BufferId]uint64{},
		buffers:       buffers.New(),
		modified:      map[buffers.BufferId]bool{},
		s:             s,
	}
}

func (e *Editor) NewBuffer(filepath string, r io.Reader) (buffers.BufferId, error) {
	buf, err := buffer.NewFromReader(r)
	if err != nil {
		log.Printf("[NewBuffer] %v\n", err)
		e.statusmsg(fmt.Sprintf("Buffer open failed: %v", err))
		return 0, err
	}
	return e.NewFromBuffer(filepath, buf)
}

func (e *Editor) setactivebuf(bid buffers.BufferId) {
	log.Println("[setactivebuf] ", bid)
	if _, ok := e.bufpopularity[bid]; !ok {
		e.bufpopularity[bid] = 1
	} else {
		e.bufpopularity[bid]++
	}
	e.activebuf = bid
}

func (e *Editor) sethighlighting() {
	eb := e.buffers.Get(e.activebuf)
	ec := config.GetEditorConfig(eb.Filepath)

	// If we have no highlights, use a fast dummy highlighter.
	if len(ec.HighlightPatterns) == 0 && len(ec.HighlightKeywords) == 0 {
		eb.SetHighlighting(highlighting.NewDummy())
		return
	}

	hl := highlighting.New(eb.Buffer.ToRunes())
	for _, pat := range ec.HighlightPatterns {
		hl.Pattern(pat.Pattern, pat.Left, pat.Right, pat.Style, uint8(pat.Priority))
	}
	for _, kw := range ec.HighlightKeywords {
		hl.Keyword(kw.Keyword, kw.Style, 1)
	}
	hl.Analyze()
	eb.SetHighlighting(hl)
}

func (e *Editor) highlightline(lineno int) {
	eb := e.buffers.Get(e.activebuf)
	eb.Hilite.ModifyLine(lineno, eb.Buffer.GetLine(lineno))
}

func (e *Editor) NewFromBuffer(filepath string, buf *buffer.Buffer) (buffers.BufferId, error) {
	bid := e.buffers.New(filepath, buf)
	e.setactivebuf(bid)
	e.sethighlighting()
	return bid, nil
}

func (e *Editor) askyesno(prompt string) bool {
	d := dialog.New(prompt)
	_, h := e.s.Size()
	for {
		key, r := d.Ask(e.s, 0, h-1)
		log.Printf("[askyesno] %s %c\n", tcell.KeyNames[key], r)
		switch {
		case key == tcell.KeyRune && (r == 'y' || r == 'Y'):
			return true
		case (key == tcell.KeyRune && (r == 'n' || r == 'N')) ||
			key == tcell.KeyCtrlC:
			return false
		}
	}
}

func (e *Editor) closeactivebuffer(force bool) bool {
	if e.ismodified() && !force {
		if !e.askyesno("Unchanged saves to buffer, close [y/n]?") {
			return false
		}
	}

	waslast := e.closebuffer(e.activebuf)
	// Now that the current buffer is closed, choose the
	// new active buffer based on its popularity.
	popbid := buffers.BufferId(0)
	votemax := uint64(0)
	for curbid, votes := range e.bufpopularity {
		if votes > votemax {
			popbid = curbid
		}
	}
	e.setactivebuf(popbid)
	return waslast
}

func (e *Editor) closebuffer(bid buffers.BufferId) bool {
	log.Printf("[closebuffer] %d\n", bid)
	e.buffers.Close(bid)
	delete(e.bufpopularity, bid)
	waslast := e.buffers.Len() == 0
	if waslast {
		e.NewFromBuffer("", buffer.New(nil))
	}
	return waslast
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

	if eb == nil {
		panic(fmt.Sprintf(
			"no activebuf when drawing, got %d -> %#v [%#v]",
			e.activebuf,
			eb,
			e.buffers.All()))
	}
	w, h := e.s.Size()
	rend := eb.Viewport.Render(w, h-1, eb.CursorLine(), eb.CursorCol(), eb.Hilite)
	col := 0
	lineno := 0
	for h > 0 && rend.Scan() {
		rl := rend.Line()
		for i, r := range rl.Content {
			e.s.SetContent(col+i, lineno, r, nil, rl.GetStyle(i))
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
	e.highlightline(eb.CursorLine())
}

func (e *Editor) insertlinefeed() {
	eb := e.buffers.Get(e.activebuf)
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewLinefeed(eb.Cursor())))
	e.highlightline(eb.CursorLine() - 1)
	eb.Hilite.InsertLine(eb.CursorLine(), eb.Buffer.GetLine(eb.CursorLine()))
}

func (e *Editor) backspaceordelword(backspace bool) {
	var act func(int, int) *buffer.Action
	switch backspace {
	case true:
		act = buffer.NewBackspace
	case false:
		act = buffer.NewDelWord
	}
	eb := e.buffers.Get(e.activebuf)
	linenobefore := eb.CursorLine()
	eb.Update(eb.Buffer.Perform(act(eb.Cursor())))
	linenoafter := eb.CursorLine()
	if linenobefore == linenoafter {
		e.highlightline(eb.CursorLine())
	} else {
		eb.Hilite.DeleteLine(linenobefore)
		e.highlightline(eb.CursorLine())
	}
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

func (e *Editor) setmodified(status bool) {
	e.modified[e.activebuf] = status
}

func (e *Editor) ismodified() bool {
	return e.modified[e.activebuf]
}

func (e *Editor) savebuffer() {
	eb := e.buffers.Get(e.activebuf)
	log.Printf("[savebuffer] %q\n", eb.Filepath)

	_, h := e.s.Size()
	fp, err := textentry.
		New(eb.Filepath, "Save: ", 512).
		Ask(e.s, 0, h-1)
	if err != nil {
		log.Println("[savebuffer, error-ask] ", err)
		return
	}
	if len(fp) == 0 {
		return
	}

	abspath, err := filepath.Abs(string(fp))
	if err != nil {
		log.Println("[savebuffer, error-abs] ", err)
		e.statusmsg(fmt.Sprintf("%v", err))
		return
	}
	if fi, err := os.Stat(string(fp)); err == nil {
		if fi.IsDir() {
			log.Println("[savebuffer, is-dir]")
			e.statusmsg(fmt.Sprintf("Cannot save, it's a directory: %s", abspath))
			return
		}
	}

	log.Println("[savebuffer, abs] ", abspath)
	if err := eb.Buffer.Save(abspath); err != nil {
		log.Println("[savebuffer] failed: ", err)
		e.statusmsg(fmt.Sprintf("Cannot save buffer: %v", err))
		return
	}
	eb.Filepath = abspath
	e.setmodified(false)

	ec := config.GetEditorConfig(eb.Filepath)
	eb.Buffer.TabSize = ec.TabSize
	if len(ec.SaveHook) == 0 {
		return
	}

	rcommand := []string{}
	for _, part := range ec.SaveHook {
		rcommand = append(
			rcommand, strings.ReplaceAll(part, "__ABSPATH__", abspath))
	}
	log.Printf("[savebuffer, hook] %#v -> %#v\n", ec.SaveHook, rcommand)
	e.execandreload(abspath, rcommand)
}

func (e *Editor) execandreload(abspath string, cmd []string) {
	c := exec.Command(cmd[0], cmd[1:]...)
	log.Printf("[execandreload, command] %#v\n", c)
	out, err := c.Output()
	log.Printf("[execandreload, output] %q\n", out)
	if err != nil {
		log.Printf("[execandreload, exec error] %v\n", err)
		e.statusmsg(fmt.Sprintf("Hook failed: %v", err))
		return
	}
	f, err := os.Open(abspath)
	if err != nil {
		log.Printf("[execandreload, reopen error] %v\n", err)
		e.statusmsg(fmt.Sprintf("Reopening buffer failed: %v", err))
		return
	}
	defer f.Close()

	oldbuf := e.buffers.Get(e.activebuf)
	oldviewportstart := oldbuf.Viewport.Start()
	oldcursorline, oldcursorcol := oldbuf.CursorLine(), oldbuf.CursorCol()

	log.Println("[execandreload, reopened] ", abspath)
	newbid, err := e.NewBuffer(abspath, f)
	if err != nil {
		log.Printf("[execandreload, newbuffer error] %v\n", err)
		return
	}
	// Now that the new buffer was opened successfully, we can get
	// rid of the old one.
	e.closebuffer(oldbuf.Id())
	e.setactivebuf(newbid)

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
		New("", "Line: ", 12).
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
	eb.Viewport.SetTeleported(eb.CursorLine())
}

func (e *Editor) delline() {
	eb := e.buffers.Get(e.activebuf)
	if eb.Buffer.Lines() == 1 && eb.Buffer.LineLength(eb.CursorLine()) == 0 {
		return
	}

	if eb.Buffer.LineLength(eb.CursorLine()) == 0 {
		eb.Update(eb.Buffer.Perform(buffer.NewDelLine(eb.CursorLine())))
		eb.Hilite.DeleteLine(eb.CursorLine() + 1)
		e.highlightline(eb.CursorLine())
		return
	}
	eb.Update(eb.Buffer.Perform(buffer.NewDelLineContent(eb.CursorLine(), eb.CursorCol())))
	e.highlightline(eb.CursorLine())
}

func (e *Editor) jumpword(left bool) {
	eb := e.buffers.Get(e.activebuf)
	eb.SetCursor(eb.Buffer.JumpWord(eb.CursorLine(), eb.CursorCol(), left))
}

func (e *Editor) replace() {
	eb := e.buffers.Get(e.activebuf)
	_, h := e.s.Size()
	from, err := textentry.
		New(e.prevsearch[eb.Id()], "Replace: ", 256).
		Ask(e.s, 0, h-1)
	if err != nil || len(from) == 0 {
		return
	}

	to, err := textentry.
		New(e.prevsearch[eb.Id()], "... with: ", 256).
		Ask(e.s, 0, h-1)
	if err != nil {
		return
	}

	log.Printf("[replace] %q -> %q\n", string(from), string(to))
	limits := &buffer.SearchLimit{
		StartLineno: eb.CursorLine(),
		StartCol:    eb.CursorCol(),
		EndLineno:   eb.Buffer.Lines() - 1,
		EndCol:      eb.Buffer.LineLength(eb.Buffer.Lines() - 1),
	}
	log.Printf("[search, limits] %#v\n", limits)
	if lineno, col := eb.Buffer.ReplaceRange(from, to, limits); lineno != -1 && col != -1 {
		log.Printf("[replace, found] (%d, %d)\n", lineno, col)
		eb.SetCursor(lineno, col)
		eb.Viewport.SetTeleported(eb.CursorLine())
		e.setmodified(true)
		e.highlightline(lineno)
	}
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
		case errors.Is(err, textentry.ErrorCancelled):
			return
		default:
			log.Println("[search, error-ask] ", err)
			e.statusmsg(fmt.Sprintf("%v", err))
			return
		}

		limits := &buffer.SearchLimit{
			StartLineno: eb.CursorLine(),
			StartCol:    prevcol,
			EndLineno:   eb.Buffer.Lines() - 1,
			EndCol:      eb.Buffer.LineLength(eb.Buffer.Lines() - 1),
		}
		log.Printf("[search, limits] %#v\n", limits)
		if lineno, col := eb.Buffer.SearchRange(term, limits); lineno != -1 && col != -1 {
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

func (e *Editor) statusmsg(msg string) {
	log.Println("[drawstatusmsg] ", msg)
	if e.s == nil {
		log.Println("[drawstatusmsg] screen not yet initialized")
		return
	}
	w, h := e.s.Size()
	m := []rune(fmt.Sprintf("<> %s [*]", msg))
	for _, msgpart := range util.SplitRunesOnWidth(m, w) {
		dialog.New(string(msgpart)).Ask(e.s, 0, h-1)
	}
}

func (e *Editor) drawstatusline() {
	w, h := e.s.Size()
	fn := "{no file}"
	lineno := 0
	col := 0
	eb := e.buffers.Get(e.activebuf)
	f := eb.Filepath
	if len(f) > 0 {
		fn = f
	}
	lineno = eb.CursorLine() + 1
	col = eb.CursorCol() + 1

	var modified rune
	if e.ismodified() {
		modified = '*'
	} else {
		modified = ' '
	}

	fn = string(util.TruncateLine([]rune(fn), w-20, ':'))
	line := []rune(
		fmt.Sprintf(
			"[%03d] %3d, %2d: %c %s", e.activebuf, lineno, col, modified, fn))
	for i, r := range line {
		e.s.SetContent(i, h-1, r, nil, config.STYLE_DEFAULT)
		if i > w {
			break
		}
	}
}

func (e *Editor) jumpempty(up bool) {
	eb := e.buffers.Get(e.activebuf)
	defer func() {
		eb.Viewport.SetTeleported(eb.CursorLine())
	}()
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
		eb.Viewport.SetTeleported(eb.CursorLine())
		// XXX We reanalyse the whole file after undo.
		e.sethighlighting()
	}
	if !eb.Buffer.Modified() {
		e.setmodified(false)
	}
}

func (e *Editor) backtab() {
	eb := e.buffers.Get(e.activebuf)
	len0 := eb.Buffer.LineLength(eb.CursorLine())
	eb.Update(
		eb.Buffer.Perform(
			buffer.NewDetabulate(eb.Cursor())))
	if len0 != eb.Buffer.LineLength(eb.CursorLine()) {
		e.setmodified(true)
		e.highlightline(eb.CursorLine())
	}
}

func (e *Editor) openbuffer() {
	var rootdir string

	if len(e.prevopendir) > 0 {
		rootdir = e.prevopendir
	} else {
		var err error
		rootdir, err = os.Getwd()
		if err != nil {
			rootdir, err = os.UserHomeDir()
			if err != nil {
				rootdir = "/"
			}
		}
	}

	w, h := e.s.Size()

	fp, err := textentry.
		New(rootdir, "Directory: ", 512).
		Ask(e.s, 0, h-1)
	if err != nil {
		log.Println("[openbuffer, error-ask] ", err)
		return
	}
	absrootdir, err := filepath.Abs(string(fp))
	if err != nil {
		log.Println("[openbuffer, error-abs] ", err)
		e.statusmsg(fmt.Sprintf("%v", err))
		return
	}
	if fi, err := os.Stat(absrootdir); err != nil {
		log.Printf("[openbuffer, stat] %q: %v\n", absrootdir, err)
		return
	} else if !fi.IsDir() {
		log.Printf("[openbuffer, notdir] %q\n", absrootdir)
		return
	}

	e.prevopendir = absrootdir
	lib := library.New()
	lib.Add(absrootdir)
	choices := []fuzzyselect.Entry{}
	paths := 0
	lib.Walk(func(abspath string) error {
		choices = append(choices, fuzzyselect.Entry{Display: []rune(abspath), Id: uint32(paths)})
		paths++
		return nil
	})

	sel, err := fuzzyselect.New(choices).Choose(e.s, 0, 0, w, h-2)
	if err != nil {
		log.Printf("[openbuffer, fuzzy error] %v\n", err)
		return
	}
	fn := string(sel.Display)
	f, err := os.Open(fn)
	if err != nil {
		log.Printf("[openbuffer, open error] %v\n", err)
		e.statusmsg(fmt.Sprintf("Opening file failed: %v", err))
		return
	}
	defer f.Close()
	if fi, err := f.Stat(); err != nil {
		log.Printf("[openbuffer, stat error] %v\n", err)
		e.statusmsg(fmt.Sprintf("Stat failed: %v", err))
		// This does not have to be a hard failure. We can
		// proceed cautiously if we managed to open the file
		// regardless of Stat failing.
	} else if fi.Size() > config.WARNFILESZ {
		log.Printf("[openbuffer, too large]: %d\n", fi.Size())
		if !e.askyesno(fmt.Sprintf(
			"%q is %d MB, do you really want to open it? [y/n]",
			fi.Name(), fi.Size()/1024/1024)) {
			return
		}
	}
	e.NewBuffer(fn, f)
	log.Printf("[openbuffer, done] %q\n", fn)
}

func (e *Editor) changebuffer() {
	choices := []fuzzyselect.Entry{}

	for bufnum, bufentry := range e.buffers.All() {
		var display string
		fp := bufentry.Filepath
		if len(fp) > 0 {
			display = fp
		} else {
			display = fmt.Sprintf("{buffer-%03d}", bufnum)
		}
		choices = append(choices, fuzzyselect.Entry{
			Display: []rune(display),
			Id:      uint32(bufnum),
		})
	}

	w, h := e.s.Size()
	sel, err := fuzzyselect.New(choices).Choose(e.s, 0, 0, w, h-2)
	if err != nil {
		log.Printf("[changebuffer, fuzzy error] %v\n", err)
		return
	}
	e.setactivebuf(buffers.BufferId(sel.Id))
}

func (e *Editor) quit() bool {
	if !e.askyesno("Quit [y/n]?") {
		return false
	}
	for {
		if e.ismodified() {
			e.s.Clear()
			e.drawactivebuf()
			e.s.Show()
			e.savebuffer()
		}
		waslast := e.closeactivebuffer(true)
		if waslast {
			break
		}
	}
	return true
}

func (e *Editor) Run() error {
	if e.s == nil {
		if err := e.initscreen(); err != nil {
			return err
		}
	}
	if e.buffers.Len() == 0 {
		e.NewFromBuffer("", buffer.New(nil))
	}
	e.s.Clear()
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
			case ev.Key() == tcell.KeyCtrlF:
				e.openbuffer()
			case ev.Key() == tcell.KeyCtrlN:
				e.NewFromBuffer("", buffer.New(nil))
			case ev.Key() == tcell.KeyCtrlP:
				e.changebuffer()
			case ev.Key() == tcell.KeyCtrlUnderscore:
				e.undo()
			case ev.Key() == tcell.KeyCtrlR:
				e.replace()
			case ev.Key() == tcell.KeyCtrlS:
				e.search()
			case ev.Key() == tcell.KeyCtrlK:
				e.delline()
				e.setmodified(true)
			case ev.Key() == tcell.KeyCtrlG:
				e.jumpline()
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Rune() == 'f':
				e.closeactivebuffer(false)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyUp:
				e.jumpempty(true)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyDown:
				e.jumpempty(false)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyLeft:
				e.jumpword(true)
			case (ev.Modifiers()&tcell.ModAlt > 0) && ev.Key() == tcell.KeyRight:
				e.jumpword(false)
			case ev.Key() == tcell.KeyCtrlX:
				if e.quit() {
					e.s.Fini()
					break main
				}
			case ev.Key() == tcell.KeyRune:
				e.insertrune(ev.Rune())
				e.setmodified(true)
			case ev.Key() == tcell.KeyEnter:
				e.insertlinefeed()
				e.setmodified(true)
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				if ev.Modifiers()&tcell.ModAlt > 0 {
					e.backspaceordelword(false)
				} else {
					e.backspaceordelword(true)
				}
				e.setmodified(true)
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
				eb := e.buffers.Get(e.activebuf)
				c := config.GetEditorConfig(eb.Filepath)
				if c.TabSpaces {
					for i := 0; i < c.TabSize; i++ {
						e.insertrune(' ')
					}
				} else {
					e.insertrune('\t')
				}
				e.setmodified(true)
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
