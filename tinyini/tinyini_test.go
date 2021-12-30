package tinyini_test

import (
	"fmt"
	"strings"
	"testing"

	tu "github.com/susji/ked/internal/testutil"
	"github.com/susji/ked/tinyini"
)

func TestBasic(t *testing.T) {
	c := `
globalkey = globalvalue
[section]
key = value
anotherkey = "  has whitespace   "
`
	res, errs := tinyini.Parse(strings.NewReader(c))
	tu.Assert(t, len(errs) == 0, "should have no error, got %v", errs)
	tu.Assert(t, res[""] != nil, "missing global section")
	tu.Assert(t, res["section"] != nil, "missing section")
	tu.Assert(t, res[""]["globalkey"] == "globalvalue", "missing global value")
	tu.Assert(t, res["section"]["key"] == "value", "missing sectioned value")
	tu.Assert(
		t,
		res["section"]["anotherkey"] == "  has whitespace   ",
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
