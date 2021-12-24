package highlighting_test

import (
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/highlighting"
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

	h := highlighting.New(msg).
		Pattern("(//.+)", 0, 1, s3).
		Keyword("of", s1).
		Keyword("words", s2).
		Keyword("notexist", sz).
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
	h := highlighting.New(msg).
		Keyword(`func`, s).
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
	h := highlighting.New(msg).
		Keyword(`func`, s).
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

func TestGreediness(t *testing.T) {
	msg := [][]rune{
		[]rune(`// first match should win`),
	}
	s1 := tcell.StyleDefault.Italic(true)
	s2 := tcell.StyleDefault.Bold(true)
	h := highlighting.New(msg).
		Pattern(`//.*`, 0, 1, s1).
		Keyword(`first`, s2).
		Analyze()

	g1 := h.Get(0, len("/")-1)
	g2 := h.Get(0, len("// fir")-1)
	g3 := h.Get(0, len("// first match should win")-1)

	tu.Assert(t, s1 == g1, "got %x, want %x", g1, s1)
	tu.Assert(t, s1 == g2, "got %x, want %x", g2, s1)
	tu.Assert(t, s1 == g3, "got %x, want %x", g3, s1)
}
