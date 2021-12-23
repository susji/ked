package highlighting

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type Highlighting struct {
	source   [][]rune
	styles   [][]tcell.Style
	mappings []Mapping
}

type Mapping struct {
	pattern       *regexp.Regexp
	style         tcell.Style
	lefti, righti int
}

func New(source [][]rune) *Highlighting {
	return &Highlighting{source: source}
}

func (h *Highlighting) Pattern(pattern string, lefti, righti int, style tcell.Style) *Highlighting {
	pat := regexp.MustCompile(pattern)
	h.mappings = append(h.mappings, Mapping{
		pattern: pat,
		style:   style,
		lefti:   lefti,
		righti:  righti,
	})
	return h
}

func (h *Highlighting) Keyword(keyword string, style tcell.Style) *Highlighting {
	pat := regexp.MustCompile(fmt.Sprintf(`([^\w]|^)(%s)([^\w]|$)`, keyword))
	h.mappings = append(h.mappings, Mapping{
		pattern: pat,
		style:   style,
		lefti:   4,
		righti:  5,
	})
	return h
}

func (h *Highlighting) Analyze() *Highlighting {
	h.styles = [][]tcell.Style{}
	for lineno, line := range h.source {
		h.styles = append(h.styles, make([]tcell.Style, len(line)))
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
					h.styles[lineno][col] = mapping.style
				}
				// We skip over this many runes due to present match.
				runeacc += right
				l = l[righti:]
			}
		}
	}
	return h
}

func (h *Highlighting) Get(lineno, col int) tcell.Style {
	if lineno >= len(h.styles) || col >= len(h.styles[lineno]) {
		return tcell.StyleDefault
	}
	return h.styles[lineno][col]
}
