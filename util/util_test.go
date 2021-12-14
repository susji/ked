package util_test

import (
	"fmt"
	"testing"

	tu "github.com/susji/ked/internal/testutil"
	"github.com/susji/ked/util"
)

func TestTruncateLine(t *testing.T) {
	table := []struct {
		give, want string
		width      int
	}{
		{"abc", "abc", 3},
		{"abcd", ".cd", 3},
		{"1234567890", ".", 1},
		{"123", "", 0},
		{"123", "123", 1000},
	}

	for _, e := range table {
		t.Run(fmt.Sprintf("%s_%s_%d", e.give, e.want, e.width), func(t *testing.T) {
			got := string(util.TruncateLine([]rune(e.give), e.width, '.'))
			tu.Assert(
				t,
				got == e.want,
				"got %q, wanted %q",
				got,
				e.want)
		})
	}
}
