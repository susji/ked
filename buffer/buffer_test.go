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
