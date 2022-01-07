package dialog

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/config"
)

type Dialog struct {
	prompt string
}

func New(prompt string) *Dialog {
	return &Dialog{prompt: prompt}
}

func (d *Dialog) draw(s tcell.Screen, prompt []rune, col, lineno int) {
	w, _ := s.Size()
	for i := col; i < w; i++ {
		s.SetContent(i, lineno, ' ', nil, config.STYLE_DEFAULT)
	}
	for i, r := range prompt {
		s.SetContent(col+i, lineno, r, nil, config.STYLE_DEFAULT)
	}
	s.ShowCursor(len(prompt), lineno)
	s.Show()
}

func (d *Dialog) Ask(s tcell.Screen, col, lineno int) (tcell.Key, rune) {
	prompt := []rune(d.prompt)
	d.draw(s, prompt, col, lineno)

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("[dialog, EventKey] %s (mods=%X)\n", ev.Name(), ev.Modifiers())
			return ev.Key(), ev.Rune()
		}
	}
}
