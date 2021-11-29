package buffer_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/susji/ked/buffer"
	ta "github.com/susji/ked/internal/testutil"
)

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
	gotlines := b.Lines()

	for i, gotline := range gotlines {
		wantline := lines[i]
		ta.Assert(t, reflect.DeepEqual(string(gotline.Get()), wantline), "unexpected got=%q, want=%q", string(gotline.Get()), wantline)
	}

}
