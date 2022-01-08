package config_test

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/config"
	tu "github.com/susji/ked/internal/testutil"
	"github.com/susji/tinyini"
)

func TestConfigBasic(t *testing.T) {
	c := map[string]tinyini.Section{
		"": tinyini.Section{
			"maxfiles":   []string{"1234"},
			"tabsize":    []string{"123"},
			"tabspaces":  []string{"yes"},
			"savehook":   []string{"one __ABSPATH__ two"},
			"ignoredir":  []string{".git", ".got"},
			"worddelims": []string{`ab\t\rc`},
		},
	}

	config.ParseConfig(c)

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
	c := map[string]tinyini.Section{
		"": tinyini.Section{
			"tabsize":   []string{"123"},
			"tabspaces": []string{"true"},
		},
		"filetype:*.abc": tinyini.Section{
			"tabsize":           []string{"404"},
			"tabspaces":         []string{"false"},
			"savehook":          []string{"three __ABSPATH__ four     five"},
			"highlight-keyword": []string{"dim:keyword"},
			"highlight-pattern": []string{"1:2:3:dim:pattern"},
		},
	}

	config.ParseConfig(c)
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
