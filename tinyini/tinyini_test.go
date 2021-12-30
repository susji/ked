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
`

	res, err := tinyini.Parse(strings.NewReader(c))
	tu.Assert(t, err == nil, "should have no error, got %v", err)
	tu.Assert(t, res[""]["globalkey"] == "globalvalue", "missing global value")
	tu.Assert(t, res["section"]["key"] == "value", "missing sectioned value")
}
