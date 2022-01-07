package textentry

import (
	"errors"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/config"
	"github.com/susji/ked/util"
)

var (
	ErrorCancelled = errors.New("Text entry was cancelled by user")
	ErrorTooLong   = errors.New("Too much text entry")
)

type TextEntry struct {
	defval, prompt string
	maxlen         int
	binds          []bind
}

func New(defval, prompt string, maxlen int) *TextEntry {
	return &TextEntry{
		defval: defval,
		prompt: prompt,
		maxlen: maxlen,
	}
}

func (te *TextEntry) draw(s tcell.Screen, prompt, answer []rune, col, lineno int) {
	// Yes, this is pretty hacky but in practice it will be robust
	// enough.
	w, _ := s.Size()
	answer = util.TruncateLine(answer, w-col-len(prompt), ':')
	all := append(prompt, answer...)
	for i := col; i < w; i++ {
		s.SetContent(i, lineno, ' ', nil, config.STYLE_DEFAULT)
	}
	for i, r := range all {
		s.SetContent(col+i, lineno, r, nil, config.STYLE_DEFAULT)
	}
	s.ShowCursor(col+len(all), lineno)
	s.Show()
}

func (te *TextEntry) AddBinding(key tcell.Key, reterr error) *TextEntry {
	te.binds = append(te.binds, bind{
		key:    key,
		reterr: reterr,
	})
	return te
}

func (te *TextEntry) Ask(s tcell.Screen, col, lineno int) (answer []rune, reterr error) {
	answer = []rune(te.defval)
	prompt := []rune(te.prompt)
	for {
		te.draw(s, prompt, answer, col, lineno)

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			// XXX Resize not handled at all.
			log.Printf("[text-entry, EventKey] %s (mods=%X)\n",
				ev.Name(), ev.Modifiers())
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[savebuffer, cancel]")
				reterr = ErrorCancelled
				return
			case ev.Key() == tcell.KeyRune && ev.Modifiers() == 0:
				answer = append(answer, ev.Rune())
			case ev.Key() == tcell.KeyEnter:
				reterr = nil
				return
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				if (ev.Modifiers() & tcell.ModAlt) > 0 {
					answer = []rune{}
				} else if len(answer) > 0 {
					answer = answer[:len(answer)-1]
				}
			default:
				for _, bind := range te.binds {
					if ev.Key() == bind.key {
						log.Printf("[textentry, custom-bind] %v\n", bind.reterr)
						reterr = bind.reterr
						return
					}
				}
			}

		}
		if len(answer) >= te.maxlen {
			reterr = ErrorTooLong
			return
		}
	}
}

type bind struct {
	key    tcell.Key
	reterr error
}
