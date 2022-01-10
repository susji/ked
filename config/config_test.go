package config_test

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/config"
	tu "github.com/susji/ked/internal/testutil"
	ti "github.com/susji/tinyini"
)

func TestConfigBasic(t *testing.T) {
	c := map[string]ti.Section{
		"": ti.Section{
			"maxfiles":  []ti.Pair{ti.Pair{Value: "1234", Lineno: 1}},
			"tabsize":   []ti.Pair{ti.Pair{Value: "123", Lineno: 2}},
			"tabspaces": []ti.Pair{ti.Pair{Value: "yes", Lineno: 3}},
			"savehook":  []ti.Pair{ti.Pair{Value: "one __ABSPATH__ two", Lineno: 4}},
			"ignoredir": []ti.Pair{
				ti.Pair{Value: ".git", Lineno: 5}, ti.Pair{Value: ".got", Lineno: 6}},
			"worddelims": []ti.Pair{ti.Pair{Value: `ab\t\rc`, Lineno: 7}},
		},
	}

	config.ParseConfig("test.ini", c)

	tu.Assert(t, config.MAXFILES == 1234, "unexpected maxfiles, got %d", config.MAXFILES)

	gec := config.GetEditorConfig("")

	tu.Assert(t, gec.TabSize == 123, "unexpected tabsize, got %d", gec.TabSize)
	tu.Assert(t, gec.TabSpaces, "unexpected tabspaces, got %t", gec.TabSpaces)
	tu.Assert(
		t,
		reflect.DeepEqual(gec.SaveHook, []string{"one", "__ABSPATH__", "two"}),
		"unexpected savehook: %#v",
		gec.SaveHook)
	tu.Assert(
		t,
		reflect.DeepEqual(config.IGNOREDIRS, map[string]bool{".git": true, ".got": true}),
		"unexpected ignoredirS: %#v",
		config.IGNOREDIRS)
	tu.Assert(t, config.WORD_DELIMS == "ab\t\rc", "unexpect word delims: %q", config.WORD_DELIMS)
}

func TestConfigSection(t *testing.T) {
	c := map[string]ti.Section{
		"": ti.Section{
			"tabsize":   []ti.Pair{ti.Pair{Value: "123", Lineno: 1}},
			"tabspaces": []ti.Pair{ti.Pair{Value: "true", Lineno: 2}},
		},
		"filetype:*.abc": ti.Section{
			"tabsize":           []ti.Pair{ti.Pair{Value: "404", Lineno: 3}},
			"tabspaces":         []ti.Pair{ti.Pair{Value: "false", Lineno: 4}},
			"savehook":          []ti.Pair{ti.Pair{Value: "three __ABSPATH__ four     five", Lineno: 5}},
			"highlight-keyword": []ti.Pair{ti.Pair{Value: "dim:keyword", Lineno: 6}},
			"highlight-pattern": []ti.Pair{ti.Pair{Value: "1:2:3:dim:pattern", Lineno: 7}},
		},
	}

	config.ParseConfig("test.ini", c)
	ec := config.GetEditorConfig("file.abc")
	tu.Assert(t, ec.TabSize == 404, "unexpected tabsize, got %d", ec.TabSize)
	tu.Assert(t, !ec.TabSpaces, "unexpected tabspaces, got %t", ec.TabSpaces)
	tu.Assert(
		t,
		reflect.DeepEqual(ec.SaveHook, []string{"three", "__ABSPATH__", "four", "five"}),
		"unexpected savehook, got %#v",
		ec.SaveHook)
	tu.Assert(
		t,
		reflect.DeepEqual(
			ec.HighlightKeywords,
			[]config.HighlightKeyword{
				config.HighlightKeyword{
					Keyword: "keyword",
					Style:   tcell.StyleDefault.Dim(true),
				},
			}),
		"unexpected highlight keywords: %#v",
		ec.HighlightKeywords)
	tu.Assert(
		t,
		reflect.DeepEqual(
			ec.HighlightPatterns,
			[]config.HighlightPattern{
				config.HighlightPattern{
					Priority: 1,
					Left:     2,
					Right:    3,
					Pattern:  "pattern",
					Style:    tcell.StyleDefault.Dim(true),
				},
			}),
		"unexpected highlight patterns: %#v",
		ec.HighlightPatterns)
}
