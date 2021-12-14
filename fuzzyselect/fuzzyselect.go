package fuzzyselect

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/util"
)

var (
	ErrorCancelled = errors.New("fuzzyselect cancelled")
	ErrorNoMatch   = errors.New("no matches to return")

	splitter = regexp.MustCompile(" +")
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
	filters := splitter.Split(strings.ToLower(with), -1)
	ret := []Entry{}
	for _, entry := range f.choices {
		text := strings.ToLower(string(entry.Display))

		count := 0
		for _, f := range filters {
			if strings.Contains(text, f) {
				count++
			}
		}
		if count == len(filters) {
			ret = append(ret, entry)
		}
	}
	return ret
}

func (f *FuzzySelect) drawfilter(s tcell.Screen, filter string, lineno, col, w, h int) {
	rs := []rune(filter)
	s.SetContent(col, lineno, '/', nil, tcell.StyleDefault.Bold(true))
	rs = util.TruncateLine(rs, w-1, '|')
	for curcol, r := range rs {
		x := col + curcol + 1
		y := lineno
		s.SetContent(x, y, r, nil, tcell.StyleDefault.Bold(true))
	}
	s.ShowCursor(len(filter)+1, lineno)
}

func (f *FuzzySelect) drawdata(s tcell.Screen, data []Entry, choice, lineno, col, w, h int) {
	for nentry, curentry := range data {
		if nentry >= h-lineno {
			break
		}

		st := tcell.StyleDefault
		if nentry == choice {
			st = st.Bold(true)
		}

		r := '|'
		if nentry == choice {
			r = '>'
		}
		s.SetContent(col, lineno+nentry, r, nil, st)
		rs := util.TruncateLine(curentry.Display, w-1, '|')
		for curcol, r := range rs {
			x := col + curcol + 1
			y := lineno + nentry
			if x > w {
				break
			}
			s.SetContent(x, y, r, nil, st)
		}
	}
}

func (f *FuzzySelect) Choose(s tcell.Screen, lineno, col, w, h int) (*Entry, error) {
	filter := ""
	choice := 0
	for {
		s.Clear()
		filtered := f.filter(filter)
		f.drawfilter(s, string(filter), lineno, col, w, h)
		f.drawdata(s, filtered, choice, lineno+1, col, w, h)
		s.Show()
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("[fuzzyselect, EventKey] %s (mods=%X)\n",
				ev.Name(), ev.Modifiers())
			switch {
			case ev.Key() == tcell.KeyUp:
				if choice > 0 {
					choice--
				}
			case ev.Key() == tcell.KeyDown:
				if choice < len(filtered)-1 {
					choice++
				}
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[fuzzyselect, cancel]")
				return nil, ErrorCancelled
			case ev.Key() == tcell.KeyEnter:
				if len(filtered) == 0 {
					return nil, ErrorNoMatch
				}
				entry := &filtered[choice]
				return entry, nil
			case ev.Key() == tcell.KeyBackspace, ev.Key() == tcell.KeyBackspace2:
				if (ev.Modifiers() & tcell.ModAlt) > 0 {
					filter = ""
				} else if len(filter) > 0 {
					filter = filter[:len(filter)-1]
				}
				choice = 0
			case ev.Key() == tcell.KeyRune && ev.Modifiers() == 0:
				filter += string(ev.Rune())
				choice = 0
			}
		}
	}
}
