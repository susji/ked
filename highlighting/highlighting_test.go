package highlighting_test

import (
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	hl "github.com/susji/ked/highlighting"
	tu "github.com/susji/ked/internal/testutil"
)

func TestBasic(t *testing.T) {
	msg := [][]rune{
		[]rune("öäåöäåöäå some of these words are highlighted // the end of öööö"),
	}
	s1 := tcell.StyleDefault.Bold(true)
	s2 := tcell.StyleDefault.Underline(true)
	sz := tcell.StyleDefault.Italic(true)
	s3 := tcell.StyleDefault.StrikeThrough(true)

	h := hl.New(msg).
		Pattern("(//.+)", 0, 1, s3, 255).
		Keyword("of", s1, 1).
		Keyword("words", s2, 1).
		Keyword("notexist", sz, 1).
		Analyze()

	g0 := h.Get(0, 0)
	g1 := h.Get(0, utf8.RuneCountInString("öäåöäåöäå some o"))
	g2 := h.Get(0, utf8.RuneCountInString("öäåöäåöäå some of these wor"))
	g3 := h.Get(0, len(msg[0])-1)

	tu.Assert(t, tcell.StyleDefault == g0, "want styledefault, got %x", g0)
	tu.Assert(t, s1 == g1, "got %x, want %x", g1, s1)
	tu.Assert(t, s2 == g2, "got %x, want %x", g2, s2)
	tu.Assert(t, s3 == g3, "got %x, want %x", g3, s3)
}

func TestSequential(t *testing.T) {
	msg := [][]rune{
		[]rune("func func func notfunc"),
	}
	s0 := tcell.StyleDefault
	s := tcell.StyleDefault.Bold(true)
	h := hl.New(msg).
		Keyword(`func`, s, 1).
		Analyze()

	g1 := h.Get(0, 0)
	g2 := h.Get(0, len("func fu")-1)
	g3 := h.Get(0, len("func func fu")-1)
	g4 := h.Get(0, len("func func func notfu")-1)

	tu.Assert(t, s == g1, "got %x, want %x", g1, s)
	tu.Assert(t, s == g2, "got %x, want %x", g2, s)
	tu.Assert(t, s == g3, "got %x, want %x", g3, s)
	tu.Assert(t, s0 == g4, "got %x, want %x", g4, s0)
}

func TestNoSeparation(t *testing.T) {
	msg := [][]rune{
		[]rune("func funcfuncfunc func"),
	}
	s0 := tcell.StyleDefault
	s := tcell.StyleDefault.Bold(true)
	h := hl.New(msg).
		Keyword(`func`, s, 1).
		Analyze()

	g1 := h.Get(0, len("fu")-1)
	g2 := h.Get(0, len("func fu")-1)
	g3 := h.Get(0, len("func funcfu")-1)
	g4 := h.Get(0, len("func funcfuncfunc fun")-1)

	tu.Assert(t, s == g1, "got %x, want %x", g1, s)
	tu.Assert(t, s0 == g2, "got %x, want %x", g2, s0)
	tu.Assert(t, s0 == g3, "got %x, want %x", g3, s0)
	tu.Assert(t, s == g4, "got %x, want %x", g4, s)
}

func TestMixedPriorities(t *testing.T) {
	msg := [][]rune{
		[]rune(`"word // comment" outside`),
		[]rune(`// "word"`),
	}
	scomment := tcell.StyleDefault.Italic(true)
	squoted := tcell.StyleDefault.Bold(true)

	h := hl.New(msg).
		Pattern(`[^\\]?("(.*?)([^\\]?"))`, 2, 3, squoted, 1).
		Pattern(`//.*`, 0, 1, scomment, 1).
		Analyze()

	// First line.
	g1 := h.Get(0, len(`"word`)-1)
	g2 := h.Get(0, len(`"word // com`)-1)
	g3 := h.Get(0, len(`"word // comment"`)-1)
	g4 := h.Get(0, len(`"word // comment" outsi`)-1)

	tu.Assert(t, squoted == g1, "got %x, want %x", g1, squoted)
	tu.Assert(t, squoted == g2, "got %x, want %x", g2, squoted)
	tu.Assert(t, squoted == g3, "got %x, want %x", g3, squoted)
	tu.Assert(t, hl.STYLE_DEFAULT == g4, "got %x, want %x", g4, hl.STYLE_DEFAULT)

	// Second line
	g5 := h.Get(1, len(`//`)-1)
	g6 := h.Get(1, len(`// "wo`)-1)
	g7 := h.Get(1, len(`// "word"`)-1)

	tu.Assert(t, scomment == g5, "got %x, want %x", g5, scomment)
	tu.Assert(t, scomment == g6, "got %x, want %x", g6, scomment)
	tu.Assert(t, scomment == g7, "got %x, want %x", g7, scomment)
}
