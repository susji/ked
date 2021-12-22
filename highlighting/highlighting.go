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
	pattern *regexp.Regexp
	style   tcell.Style
}

func New(source [][]rune) *Highlighting {
	return &Highlighting{source: source}
}

func (h *Highlighting) Mapping(pattern string, style tcell.Style) *Highlighting {
	pat := regexp.MustCompile(fmt.Sprintf(`(\s+|^)(%s)(\s+|$)`, pattern))
	h.mappings = append(h.mappings, Mapping{pat, style})
	return h
}

func (h *Highlighting) Analyze() *Highlighting {
	h.styles = [][]tcell.Style{}
	for lineno, line := range h.source {
		h.styles = append(h.styles, make([]tcell.Style, len(line)))
		for _, mapping := range h.mappings {
			l := string(line)
			aix := mapping.pattern.FindAllStringSubmatchIndex(l, -1)
			for _, ix := range aix {
				left := utf8.RuneCountInString(l[:ix[4]])
				right := utf8.RuneCountInString(l[:ix[5]])
				for col := left; col < right; col++ {
					h.styles[lineno][col] = mapping.style
				}
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
