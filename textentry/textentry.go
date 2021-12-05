package textentry

import (
	"errors"
	"log"

	"github.com/gdamore/tcell/v2"
)

var (
	ErrorCancelled = errors.New("Text entry was cancelled by user")
	ErrorTooLong   = errors.New("Too much text entry")
)

type TextEntry struct {
	defval, prompt string
	maxlen         int
}

func New(defval, prompt string, maxlen int) *TextEntry {
	return &TextEntry{
		defval: defval,
		prompt: prompt,
		maxlen: maxlen,
	}
}

func (te *TextEntry) draw(s tcell.Screen, what []rune, col, lineno int) {
	// Yes, this is pretty hacky but in practice it will be robust
	// enough.
	w, _ := s.Size()
	for i := col; i < w; i++ {
		s.SetContent(i, lineno, ' ', nil, tcell.StyleDefault)
	}
	for i, r := range what {
		s.SetContent(col+i, lineno, r, nil, tcell.StyleDefault)
	}
	s.ShowCursor(col+len(what), lineno)
	s.Show()
}

func (te *TextEntry) Ask(s tcell.Screen, col, lineno int) (answer []rune, reterr error) {
	answer = []rune(te.defval)
	prompt := []rune(te.prompt)
	for {
		all := append(prompt, answer...)
		te.draw(s, all, col, lineno)

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			// XXX Resize not handled at all.
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[savebuffer, cancel]")
				reterr = ErrorCancelled
				return
			case ev.Key() == tcell.KeyRune:
				answer = append(answer, ev.Rune())
			case ev.Key() == tcell.KeyEnter:
				reterr = nil
				return
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				if len(answer) > 0 {
					answer = answer[:len(answer)-1]
				}
			}

		}
		if len(answer) >= te.maxlen {
			reterr = ErrorTooLong
			return
		}
	}
}
