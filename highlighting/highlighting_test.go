package highlighting_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/highlighting"
	tu "github.com/susji/ked/internal/testutil"
)

func TestBasic(t *testing.T) {
	msg := [][]rune{
		[]rune("some of these words are highlighted // the end of öööö"),
	}
	s1 := tcell.StyleDefault.Bold(true)
	s2 := tcell.StyleDefault.Underline(true)
	sz := tcell.StyleDefault.Italic(true)
	s3 := tcell.StyleDefault.StrikeThrough(true)

	h := highlighting.New(msg).
		Mapping("of", s1).
		Mapping("words", s2).
		Mapping("notexist", sz).
		Mapping("//.+", s3).
		Analyze()

	g0 := h.Get(0, 0)
	g1 := h.Get(0, len("some o"))
	g2 := h.Get(0, len("some of these word"))
	g3 := h.Get(0, len(msg[0])-1)

	tu.Assert(t, tcell.StyleDefault == g0, "want styledefault, got %x", g0)
	tu.Assert(t, s1 == g1, "got %x, want %x", g1, s1)
	tu.Assert(t, s2 == g2, "got %x, want %x", g2, s2)
	tu.Assert(t, s3 == g3, "got %x, want %x", g3, s3)
}
