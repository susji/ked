package config_test

import (
	"reflect"
	"testing"

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
