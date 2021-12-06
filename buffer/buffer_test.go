package buffer_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/susji/ked/buffer"
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
	b.InsertRunes(0, 0, msg)

	gotlines := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines, wantlines),
		"should have updated contents, got %q", gotlines)

	//
	// Insert another line of text.
	//
	msg2 := []rune("We have some more text incoming!")
	wantlines2 := [][]rune{msg, msg2}
	b.NewLine(1)
	b.InsertRunes(1, 0, msg2)

	gotlines2 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines2, wantlines2),
		"unexpected line contents: %q", gotlines2)

	//
	// Delete first line and make sure we have the second still.
	//
	b.DeleteLine(0)
	wantlines3 := [][]rune{msg2}

	gotlines3 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines3, wantlines3),
		"unexpected line contents: %q", gotlines3)

	//
	// Delete second line and make sure the buffer is empty again.
	//
	b.DeleteLine(0)
	wantlines4 := [][]rune{}

	gotlines4 := bufferToRunes(b)
	ta.Assert(t, reflect.DeepEqual(gotlines4, wantlines4),
		"should be empty, got %q", gotlines4)
}

func TestInsertRune(t *testing.T) {
	b := buffer.New([][]rune{[]rune("There is text.")})

	b.InsertRune(0, 9, 's')
	b.InsertRune(0, 10, 'o')
	b.InsertRune(0, 11, 'm')
	b.InsertRune(0, 12, 'e')
	b.InsertRune(0, 13, ' ')

	b.InsertRune(0, 18, ' ')
	b.InsertRune(0, 19, 'h')
	b.InsertRune(0, 20, 'e')
	b.InsertRune(0, 21, 'r')
	b.InsertRune(0, 22, 'e')

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
		b.Backspace(0, 17-i)
	}
	got := b.GetLine(0)
	ta.Assert(t, reflect.DeepEqual(got, []rune("This is the line with too much text.")),
		"unexpected first line: %q", string(got))

	// Remove 'also' and 'way'.
	for i := 0; i < len(" way"); i++ {
		b.Backspace(1, 37-i)
	}
	for i := 0; i < len(" also"); i++ {
		b.Backspace(1, 29-i)
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

	b.InsertLinefeed(0, 11)

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
	b.DeleteLineContent(1, 6)
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
	b.DeleteLineContent(1, 0)
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
	b.DeleteLineContent(1, 0)
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
