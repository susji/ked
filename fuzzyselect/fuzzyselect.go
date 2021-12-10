package fuzzyselect

import (
	"errors"
	"log"

	"github.com/gdamore/tcell/v2"
)

var (
	ErrorCancelled = errors.New("fuzzyselect cancelled")
)

type Entry struct {
	Display []rune
	Id      uint32
}

type FuzzySelect struct {
	choices []Entry
}

func New(choices []Entry) *FuzzySelect {
	log.Printf("[fuzzyselect, new] %d choices\n", len(choices))
	return &FuzzySelect{choices}
}

func (f *FuzzySelect) filter(with string) []Entry {
	return f.choices
}

func (f *FuzzySelect) drawfilter(s tcell.Screen, filter string, lineno, col, w, h int) {
	rs := []rune(filter)
	s.SetContent(col, lineno, '>', nil, tcell.StyleDefault)
	for curcol := 0; curcol < w; curcol++ {
		x := col + curcol + 1
		y := lineno
		if curcol < len(rs) {
			s.SetContent(x, y, rs[curcol], nil, tcell.StyleDefault)
		} else {
			s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
		}
	}
	s.ShowCursor(len(filter)+1, lineno)
}

func (f *FuzzySelect) drawdata(s tcell.Screen, filter string, lineno, col, w, h int) {
	data := f.filter(filter)

	for nentry, curentry := range data {
		if nentry >= h-lineno {
			break
		}
		for curcol := 0; curcol < w; curcol++ {
			x := col + curcol
			y := lineno + nentry
			if curcol < len(curentry.Display) {
				s.SetContent(x, y, curentry.Display[curcol], nil, tcell.StyleDefault)
			} else {
				s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
			}
		}
	}
}

func (f *FuzzySelect) Choose(s tcell.Screen, lineno, col, w, h int) (string, error) {
	filter := []rune("")
	selection := ""
	for {
		f.drawfilter(s, string(filter), lineno, col, w, h)
		f.drawdata(s, string(filter), lineno+1, col, w, h)
		s.Show()
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("[fuzzyselect, EventKey] %s (mods=%X)\n",
				ev.Name(), ev.Modifiers())
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[fuzzyselect, cancel]")
				return selection, ErrorCancelled
			case ev.Key() == tcell.KeyEnter:
				return selection, nil
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				if (ev.Modifiers() & tcell.ModAlt) > 0 {
					filter = []rune{}
				} else if len(filter) > 0 {
					filter = filter[:len(filter)-1]
				}
			case ev.Key() == tcell.KeyRune && ev.Modifiers() == 0:
				filter = append(filter, ev.Rune())
			}
		}
	}
}
