package buffer_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/susji/ked/buffer"
	"github.com/susji/ked/config"
	ta "github.com/susji/ked/internal/testutil"
)

func bufferToRunes(buf *buffer.Buffer) [][]rune {
	ret := [][]rune{}
	for lineno := 0; lineno < buf.Lines(); lineno++ {
		ret = append(ret, buf.GetLine(lineno))
	}
	return ret
}

func TestSanity(t *testing.T) {
	lines := strings.Split(`Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore
magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea
commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident,
sunt in culpa qui officia deserunt mollit anim id est
laborum.
`, "\n")
	runes := [][]rune{}
	for _, line := range lines {
		runes = append(runes, []rune(line))
	}
	b := buffer.New(runes)
	for lineno := 0; lineno < b.Lines(); lineno++ {
		gotline := b.GetLine(lineno)
		wantline := lines[lineno]
		ta.Assert(t, reflect.DeepEqual(string(gotline), wantline),
			"unexpected got=%q, want=%q", string(gotline), wantline)
	}
}

func TestInsertDelete(t *testing.T) {
	b := buffer.New(nil)

	ta.Assert(t, b.Lines() == 0, "should have zero lines")

	//
	// Insert one line and keep it empty.
	//
	b.NewLine(0)
	ta.Assert(t, b.Lines() == 1, "should have one line")
	want := []rune{}
	line := b.GetLine(0)
	ta.Assert(t, reflect.DeepEqual(line, want), "should be empty")

	//
	// Insert some text into our line.
	//
	msg := []rune("these are the new line contents!")
	wantlines := [][]rune{msg}
	b.Perform(buffer.NewInsert(0, 0, msg))

	gotlines := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines, wantlines),
		"should have updated contents, got %q", gotlines)

	//
	// Insert another line of text.
	//
	msg2 := []rune("We have some more text incoming!")
	wantlines2 := [][]rune{msg, msg2}
	b.NewLine(1)
	b.Perform(buffer.NewInsert(1, 0, msg2))

	gotlines2 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines2, wantlines2),
		"unexpected line contents: %q", gotlines2)

	//
	// Delete first line and make sure we have the second still.
	//
	b.Perform(buffer.NewDelLine(0))
	wantlines3 := [][]rune{msg2}

	gotlines3 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines3, wantlines3),
		"unexpected line contents: %q", gotlines3)

	//
	// Delete second line and make sure the buffer is empty again.
	//
	b.Perform(buffer.NewDelLine(0))
	wantlines4 := [][]rune{}

	gotlines4 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines4, wantlines4),
		"should be empty, got %q", gotlines4)
}

func TestInsertRune(t *testing.T) {
	b := buffer.New([][]rune{[]rune("There is text.")})

	b.Perform(buffer.NewInsert(0, 9, []rune{'s'}))
	b.Perform(buffer.NewInsert(0, 10, []rune{'o'}))
	b.Perform(buffer.NewInsert(0, 11, []rune{'m'}))
	b.Perform(buffer.NewInsert(0, 12, []rune{'e'}))
	b.Perform(buffer.NewInsert(0, 13, []rune{' '}))

	b.Perform(buffer.NewInsert(0, 18, []rune{' '}))
	b.Perform(buffer.NewInsert(0, 19, []rune{'h'}))
	b.Perform(buffer.NewInsert(0, 20, []rune{'e'}))
	b.Perform(buffer.NewInsert(0, 21, []rune{'r'}))
	b.Perform(buffer.NewInsert(0, 22, []rune{'e'}))

	got := b.GetLine(0)
	want := []rune("There is some text here.")
	ta.Assert(t, reflect.DeepEqual(got, want), "unexpected line: %q", string(got))
}

func TestBackspace(t *testing.T) {
	b := buffer.New([][]rune{
		[]rune("This is the first line with too much text."),
		[]rune("However, the second line also has way too many runes!"),
	})

	// Remove 'first'.
	for i := 0; i < len(" first"); i++ {
		b.Perform(buffer.NewBackspace(0, 17-i))
	}
	got := b.GetLine(0)
	ta.Assert(t, reflect.DeepEqual(got, []rune("This is the line with too much text.")),
		"unexpected first line: %q", string(got))

	// Remove 'also' and 'way'.
	for i := 0; i < len(" way"); i++ {
		b.Perform(buffer.NewBackspace(1, 37-i))
	}
	for i := 0; i < len(" also"); i++ {
		b.Perform(buffer.NewBackspace(1, 29-i))
	}
	got = b.GetLine(1)
	ta.Assert(t, reflect.DeepEqual(got, []rune("However, the second line has too many runes!")),
		"unexpected second line: %q", string(got))
}

func TestLinefeed(t *testing.T) {
	b := buffer.New([][]rune{
		[]rune("First line.Second line."),
		[]rune("Third line."),
	})

	b.Perform(buffer.NewLinefeed(0, 11))

	wants := [][]rune{
		[]rune("First line."),
		[]rune("Second line."),
		[]rune("Third line."),
	}

	for i, want := range wants {
		got := b.GetLine(i)
		ta.Assert(t, reflect.DeepEqual(got, want),
			"unexpected got: %q, want: %q", string(got), string(want))
	}
}

func TestDeleteLineContent(t *testing.T) {
	b := buffer.New([][]rune{
		[]rune("First line."),
		[]rune("Second line."),
		[]rune("Third line."),
	})

	// First delete middle line partially
	b.Perform(buffer.NewDelLineContent(1, 6))
	wants1 := [][]rune{
		[]rune("First line."),
		[]rune("Second"),
		[]rune("Third line."),
	}

	for i, want := range wants1 {
		got := b.GetLine(i)
		ta.Assert(t, reflect.DeepEqual(got, want),
			"unexpected got: %q, want: %q", string(got), string(want))
	}

	// Then delete rest of middle.
	b.Perform(buffer.NewDelLineContent(1, 0))
	wants2 := [][]rune{
		[]rune("First line."),
		[]rune(""),
		[]rune("Third line."),
	}

	for i, want := range wants2 {
		got := b.GetLine(i)
		ta.Assert(t, reflect.DeepEqual(got, want),
			"unexpected got: %q, want: %q", string(got), string(want))
	}

	// Then delete empty middle line.
	b.Perform(buffer.NewDelLine(1))
	wants3 := [][]rune{
		[]rune("First line."),
		[]rune("Third line."),
	}

	for i, want := range wants3 {
		got := b.GetLine(i)
		ta.Assert(t, reflect.DeepEqual(got, want),
			"unexpected got: %q, want: %q", string(got), string(want))
	}

}

func TestSearch(t *testing.T) {
	b := buffer.New([][]rune{
		[]rune("First line."),
		[]rune("Second line."),
		[]rune("Third line."),
	})

	lineno, col := b.Search([]rune(strings.ToLower("First")))
	ta.Assert(t, lineno == 0 && col == 0, "unexpected lineno & col: %d, %d", lineno, col)

	lineno, col = b.Search([]rune(" line."))
	ta.Assert(t, lineno == 0 && col == 5, "unexpected lineno & col: %d, %d", lineno, col)

	lineno, col = b.Search([]rune("second"))
	ta.Assert(t, lineno == 1 && col == 0, "unexpected lineno & col: %d, %d", lineno, col)

	limits := &buffer.SearchLimit{
		StartLineno: 2,
		StartCol:    0,
		EndLineno:   2,
		EndCol:      len([]rune("Third line.")) - 1,
	}
	lineno, col = b.SearchRange([]rune("line"), limits)
	ta.Assert(t, lineno == 2 && col == 6, "unexpected lineno & col: %d, %d", lineno, col)

	lineno, col = b.Search([]rune("nonexistent"))
	ta.Assert(t, lineno == -1 && col == -1, "unexpected lineno & col: %d, %d", lineno, col)
}

func TestNextPrevRune(t *testing.T) {
	b := buffer.New([][]rune{
		[]rune("Pretty"),
		[]rune("short"),
		[]rune("lines"),
	})

	table := []struct {
		f           func(int, int) (rune, error)
		lineno, col int
		want        rune
	}{
		{b.NextRune, 1, 0, 'h'},
		{b.NextRune, 0, len("Pretty") - 1, 's'},
		{b.NextRune, 0, 0, 'r'},
		{b.NextRune, 1, len("short") - 1, 'l'},
		{b.NextRune, 2, len("li") - 1, 'n'},
		{b.PrevRune, 0, 1, 'P'},
		{b.PrevRune, 0, len("Pretty") - 1, 't'},
		{b.PrevRune, 2, 0, 't'},
		{b.PrevRune, 2, len("lin") - 1, 'i'},
	}

	for i, entry := range table {
		t.Run(fmt.Sprintf("%d_(%d,%d)_%c", i, entry.lineno, entry.col, entry.want),
			func(t *testing.T) {
				got, err := entry.f(entry.lineno, entry.col)
				if err != nil {
					t.Fatal("should not error but did: ", err)
				}
				want := entry.want
				ta.Assert(t, got == want, "got rune %c, want %c", got, want)
			})
	}
}

func TestJump(t *testing.T) {
	msg := [][]rune{
		[]rune("First line with some text and a"),
		[]rune("second line with more runes"),
		[]rune("  flow into a third line with the end."),
	}
	b := buffer.New(msg)

	// Jump left from the middle of a line.
	lineno, col := b.JumpWord(1, 13, true)
	ta.Assert(t, lineno == 1 && col == 12, "unexpected jump pos: %d, %d", lineno, col)

	// Jump left from the beginning of a line.
	lineno, col = b.JumpWord(1, 0, true)
	ta.Assert(t, lineno == 0 && col == 30, "unexpected jump pos: %d, %d", lineno, col)

	// Jump right from the middle of a line.
	lineno, col = b.JumpWord(1, 13, false)
	ta.Assert(t, lineno == 1 && col == 17, "unexpected jump pos: %d, %d", lineno, col)

	// Jump right from the end of a line.
	lineno, col = b.JumpWord(1, len(msg[1])-1, false)
	ta.Assert(t, lineno == 2 && col == 2, "unexpected jump pos: %d, %d", lineno, col)

	// Grande finale: Try jumping right towards the lonely 'a'.
	lineno, col = b.JumpWord(0, 26, false)
	ta.Assert(t, lineno == 0 && col == len(msg[0])-1, "unexpected jump pos: %d, %d", lineno, col)
}

func TestUndo(t *testing.T) {
	msg := [][]rune{
		[]rune("First line with some text and a"),
		[]rune("second line with more runes"),
		[]rune("  flow into a third line with the end."),
	}
	add := []rune("UNDO")
	b := buffer.New(msg)

	//
	// Add modification and verify it's there.
	//
	b.Perform(buffer.NewInsert(0, 0, add))
	ta.Assert(
		t,
		reflect.DeepEqual(b.GetLine(0), append(add, msg[0]...)),
		"unexpected after addition: %q",
		string(b.GetLine(0)))

	//
	// Undo and verify.
	//
	res := b.UndoModification()
	ta.Assert(t, res != nil, "res should not be nil")
	ta.Assert(
		t,
		reflect.DeepEqual(b.GetLine(0), msg[0]),
		"should return back to pre-undo state but did not: %q",
		string(b.GetLine(0)))

	//
	// Delete first and last lines.
	//
	b.Perform(buffer.NewDelLineContent(2, 0))
	b.Perform(buffer.NewDelLineContent(0, 0))
	b.Perform(buffer.NewDelLine(2))
	b.Perform(buffer.NewDelLine(0))
	// Verify we have only one line.
	ta.Assert(t, b.Lines() == 1, "wanted one line, got %d", b.Lines())
	ta.Assert(
		t,
		reflect.DeepEqual(b.GetLine(0), msg[1]),
		"unexpected last line: %q",
		b.GetLine(0))

	//
	// Undo and verify we're back to beginning.
	//
	for i := 0; i < 4; i++ {
		b.UndoModification()
	}
	ta.Assert(t, b.Lines() == len(msg), "wanted %d lines, got %d", len(msg), b.Lines())
	got := bufferToRunes(b)
	ta.Assert(
		t,
		reflect.DeepEqual(got, msg),
		"unexpected contents after undo, got %q",
		got)
}

func TestDetabulate(t *testing.T) {
	prefix := make([]rune, config.TABSZ)
	for i := range prefix {
		prefix[i] = ' '
	}

	msg := [][]rune{
		[]rune("\t\t\tthree tabs!"),
	}
	msg[0] = append(prefix, msg[0]...)
	b := buffer.New(msg)

	// First validate the regular spaces.
	b.Perform(buffer.NewDetabulate(0, 10))
	ta.Assert(
		t,
		reflect.DeepEqual(b.GetLine(0), msg[0][config.TABSZ:]),
		"unexpected regular space removal result :%q",
		b.GetLine(0),
	)

	// ... then the tabs.
	for i := 0; i < 3; i++ {
		b.Perform(buffer.NewDetabulate(0, 6))
		got := b.GetLine(0)
		want := msg[0][len(prefix)+i+1:]
		ta.Assert(
			t,
			reflect.DeepEqual(got, want),
			"unexpected detabulation result, got %q, want %q",
			string(got),
			string(want))
	}
}
