package util_test

import (
	"fmt"
	"reflect"
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

func TestSplitRunes(t *testing.T) {
	table := []struct {
		give  string
		want  [][]rune
		width int
	}{
		{"one two three", [][]rune{[]rune("one two three")}, 30},
		{"1112223", [][]rune{[]rune("111"), []rune("222"), []rune("3")}, 3},
		{"123", [][]rune{}, 0},
		{"123", [][]rune{[]rune("1"), []rune("2"), []rune("3")}, 1},
	}

	for _, e := range table {
		t.Run(fmt.Sprintf("%s_%v_%d", e.give, e.want, e.width), func(t *testing.T) {
			got := util.SplitRunesOnWidth([]rune(e.give), e.width)
			tu.Assert(
				t,
				reflect.DeepEqual(got, e.want),
				"got %#v, wanted %#v",
				got,
				e.want)
		})
	}
}
