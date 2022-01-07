package highlighting

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/config"
)

type highlight struct {
	priority uint8
	begincol uint16
	style    tcell.Style
}

type highlighter struct {
	source   [][]rune
	styles   [][]highlight
	mappings []Mapping
}

type Mapping struct {
	pattern       *regexp.Regexp
	style         tcell.Style
	lefti, righti int
	priority      uint8
}

type Highlighting interface {
	Get(lineno, col int) tcell.Style
	Analyze() Highlighting
	DeleteLine(lineno int) Highlighting
	ModifyLine(lineno int, data []rune) Highlighting
	InsertLine(lineno int, data []rune) Highlighting
	Pattern(pattern string, lefti, righti int, style tcell.Style, priority uint8) Highlighting
	Keyword(keyword string, style tcell.Style, priority uint8) Highlighting
}

func New(source [][]rune) Highlighting {
	return &highlighter{source: source}
}

type highlighterdummy struct{}

func NewDummy() Highlighting {
	return &highlighterdummy{}
}

func (h *highlighter) Pattern(
	pattern string, lefti, righti int, style tcell.Style, priority uint8) Highlighting {

	if priority == 0 {
		panic("priority cannot be zero")
	}

	pat := regexp.MustCompile(pattern)
	h.mappings = append(h.mappings, Mapping{
		pattern:  pat,
		style:    style,
		lefti:    lefti,
		righti:   righti,
		priority: priority,
	})
	return h
}

func (h *highlighter) Keyword(
	keyword string, style tcell.Style, priority uint8) Highlighting {

	if priority == 0 {
		panic("priority cannot be zero")
	}

	pat := regexp.MustCompile(fmt.Sprintf(`([^\w]|^)(%s)([^\w]|$)`, keyword))
	h.mappings = append(h.mappings, Mapping{
		pattern:  pat,
		style:    style,
		lefti:    4,
		righti:   5,
		priority: priority,
	})
	return h
}

func (h *highlighter) analyzeline(lineno int, line []rune) {
	for _, mapping := range h.mappings {
		l := string(line)
		runeacc := 0
		for len(l) > 0 {
			ix := mapping.pattern.FindStringSubmatchIndex(l)
			if ix == nil {
				break
			}
			lefti := ix[mapping.lefti]
			righti := ix[mapping.righti]
			left := utf8.RuneCountInString(l[:lefti])
			right := utf8.RuneCountInString(l[:righti])
			for col := runeacc + left; col < runeacc+right; col++ {
				prevpri := h.styles[lineno][col].priority
				prevcol := h.styles[lineno][col].begincol
				// Previous priority of zero means there has
				// been no styling yet applied. For this styling
				// to apply, it has to be the first one or
				// higher priority OR same priority beginning
				// earlier on the line.
				if prevpri == 0 || (mapping.priority >= prevpri &&
					uint16(runeacc+left) < prevcol) {
					h.styles[lineno][col] = highlight{
						style:    mapping.style,
						priority: mapping.priority,
						begincol: uint16(runeacc + left),
					}
				} else {
					break
				}
			}
			// We skip over this many runes due to present match.
			runeacc += right
			l = l[righti:]
		}
	}
}

func (h *highlighter) Analyze() Highlighting {
	h.styles = [][]highlight{}
	for lineno, line := range h.source {
		h.styles = append(h.styles, make([]highlight, len(line)))
		h.analyzeline(lineno, line)
	}
	return h
}

func (h *highlighter) DeleteLine(lineno int) Highlighting {
	copy(h.styles[lineno:], h.styles[lineno+1:])
	h.styles[len(h.styles)-1] = nil
	h.styles = h.styles[:len(h.styles)-1]

	return h
}

func (h *highlighter) ModifyLine(lineno int, newline []rune) Highlighting {
	h.styles[lineno] = make([]highlight, len(newline))
	h.analyzeline(lineno, newline)
	return h
}

func (h *highlighter) InsertLine(lineno int, newline []rune) Highlighting {
	h.styles = append(h.styles, []highlight{})
	copy(h.styles[lineno+1:], h.styles[lineno:])
	h.styles[lineno] = make([]highlight, len(newline))
	h.analyzeline(lineno, newline)
	return h
}

func (h *highlighter) Get(lineno, col int) tcell.Style {
	if lineno >= len(h.styles) || col >= len(h.styles[lineno]) {
		return config.STYLE_DEFAULT
	}
	return h.styles[lineno][col].style
}

func (h *highlighterdummy) Get(lineno, col int) tcell.Style {
	return config.STYLE_DEFAULT
}

func (h *highlighterdummy) Analyze() Highlighting {
	return h
}

func (h *highlighterdummy) DeleteLine(lineno int) Highlighting {
	return h
}

func (h *highlighterdummy) ModifyLine(lineno int, data []rune) Highlighting {
	return h
}

func (h *highlighterdummy) InsertLine(lineno int, data []rune) Highlighting {
	return h
}

func (h *highlighterdummy) Pattern(pattern string, lefti, righti int, style tcell.Style, priority uint8) Highlighting {
	return h
}

func (h *highlighterdummy) Keyword(
	keyword string, style tcell.Style, priority uint8) Highlighting {
	return h
}
