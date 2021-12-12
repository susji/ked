package config

import (
	"fmt"
	"regexp"
	"strings"
)

var SAVEHOOKS = map[string]string{}
var TABSZ = 4
var MAXFILES = 1_000_000
var WORD_DELIMS = " \t&|,./(){}[]#+*%'-:?!'\""
var IGNOREDIRS = map[string]bool{
	".git":         true,
	"node_modules": true,
	"__pycache__":  true,
}

func SetIgnoreDirs(rawdirs string) {
	IGNOREDIRS = map[string]bool{}
	for _, dir := range regexp.MustCompile(" *,+ *").Split(rawdirs, -1) {
		IGNOREDIRS[dir] = true
	}
}

func GetIgnoreDirsFlat() string {
	ret := []string{}
	for dir, _ := range IGNOREDIRS {
		ret = append(ret, dir)
	}
	return strings.Join(ret, ",")
}

func SetSaveHooks(rawsavehooks string) error {
	SAVEHOOKS = map[string]string{}
	for _, rawhook := range regexp.MustCompile(" *,+ *").Split(rawsavehooks, -1) {
		parts := strings.SplitN(rawhook, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unexpected savehook given: %q", rawhook)
		}
		SAVEHOOKS[parts[0]] = parts[1]
	}
	return nil
}
