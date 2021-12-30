package tinyini_test

import (
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
