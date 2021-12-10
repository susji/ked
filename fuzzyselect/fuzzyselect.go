package fuzzyselect

import (
	"errors"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	ErrorCancelled = errors.New("fuzzyselect cancelled")
	ErrorNoMatch   = errors.New("no matches to return")
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

func (f *FuzzySelect) filter(with []rune) []Entry {
	ret := []Entry{}
	for _, entry := range f.choices {
		text := strings.ToLower(string(entry.Display))
		want := strings.ToLower(string(with))
		if strings.Contains(text, want) {
			ret = append(ret, entry)
		}
	}
	return ret
}

func (f *FuzzySelect) drawfilter(s tcell.Screen, filter string, lineno, col, w, h int) {
	rs := []rune(filter)
	s.SetContent(col, lineno, '/', nil, tcell.StyleDefault.Bold(true))
	for curcol, r := range rs {
		x := col + curcol + 1
		y := lineno
		s.SetContent(x, y, r, nil, tcell.StyleDefault.Bold(true))
	}
	s.ShowCursor(len(filter)+1, lineno)
}

func (f *FuzzySelect) drawdata(s tcell.Screen, data []Entry, lineno, col, w, h int) {
	for nentry, curentry := range data {
		if nentry >= h-lineno {
			break
		}
		s.SetContent(col, lineno+nentry, '>', nil, tcell.StyleDefault.Bold(true))
		for curcol, r := range curentry.Display {
			x := col + curcol + 1
			y := lineno + nentry

			if x > w {
				break
			}

			s.SetContent(x, y, r, nil, tcell.StyleDefault)
		}
	}
}

func (f *FuzzySelect) Choose(s tcell.Screen, lineno, col, w, h int) (*Entry, error) {
	filter := []rune("")
	for {
		s.Clear()
		filtered := f.filter(filter)
		f.drawfilter(s, string(filter), lineno, col, w, h)
		f.drawdata(s, filtered, lineno+1, col, w, h)
		s.Show()
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("[fuzzyselect, EventKey] %s (mods=%X)\n",
				ev.Name(), ev.Modifiers())
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[fuzzyselect, cancel]")
				return nil, ErrorCancelled
			case ev.Key() == tcell.KeyEnter:
				if len(filtered) == 0 {
					return nil, ErrorNoMatch
				}
				entry := &filtered[0]
				return entry, nil
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
