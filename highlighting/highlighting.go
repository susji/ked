package highlighting

import (
	"fmt"
	"regexp"

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
			aix := mapping.pattern.FindAllStringSubmatchIndex(string(line), -1)
			for _, ix := range aix {
				for col := ix[4]; col < ix[5]; col++ {
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
