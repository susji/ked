package tinyini_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tu "github.com/susji/ked/internal/testutil"
	"github.com/susji/ked/tinyini"
)

func TestBasic(t *testing.T) {
	c := `
globalkey = globalvalue
[section]
key = first-value
key = second-value
anotherkey = "  has whitespace   "
`
	res, errs := tinyini.Parse(strings.NewReader(c))
	tu.Assert(t, len(errs) == 0, "should have no error, got %v", errs)
	tu.Assert(t, res[""] != nil, "missing global section")
	tu.Assert(t, res["section"] != nil, "missing section")
	tu.Assert(
		t,
		reflect.DeepEqual(res[""]["globalkey"], []string{"globalvalue"}),
		"unexpected global value: %#v", res[""]["globalkey"])
	tu.Assert(
		t,
		reflect.DeepEqual(res["section"]["key"], []string{"first-value", "second-value"}),
		"missing sectioned values")
	tu.Assert(
		t,
		reflect.DeepEqual(res["section"]["anotherkey"], []string{"  has whitespace   "}),
		"missing quoted value")
}

func TestError(t *testing.T) {
	table := []struct {
		conf string
		line int
	}{
		{`ok = value
error
`, 2},
		{`[section]
[another-section]
[borken
`, 3},
	}

	for _, entry := range table {
		t.Run(fmt.Sprintf("%s_%d", entry.conf, entry.line), func(t *testing.T) {
			_, errs := tinyini.Parse(strings.NewReader(entry.conf))
			if len(errs) != 1 {
				t.Errorf("expecting 1 error, got %d", len(errs))
				return
			}
			err := errs[0].(*tinyini.IniError)
			tu.Assert(
				t,
				err.Line == entry.line,
				"error line %d, wanted %d",
				err.Line,
				entry.line)
		})
	}
}
