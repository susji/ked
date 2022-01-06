package config

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/susji/tinyini"
)

type EditorConfig struct {
	TabSize   int
	TabSpaces bool
	SaveHook  []string
}

// "" is the global EditorConfig for non-specific filetypes
var defaultconfig = EditorConfig{
	TabSize:   4,
	TabSpaces: true,
	SaveHook:  nil,
}
var editorconfigs = map[string]*EditorConfig{
	"": &defaultconfig,
}

const (
	DEFAULT_TABSIZE   = 4
	DEFAULT_TABSPACES = true
)

var CONFFILES = getConfigFiles()
var WARNFILESZ = int64(10_485_760)
var MAXFILES = 50_000
var WORD_DELIMS = " \t=&|,./(){}[]#+*%'-:?!'\""
var IGNOREDIRS = map[string]bool{
	".git":         true,
	"node_modules": true,
	"__pycache__":  true,
	".cache":       true,
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

func splitsavehook(raw string) []string {
	return regexp.MustCompile(" +").Split(raw, -1)
}

func SetConfigFile(fn string) {
	if len(fn) == 0 {
		return
	}
	CONFFILES = []string{fn}
}

func confbool(val string) bool {
	switch strings.ToLower(val) {
	case "yes", "true", "1":
		return true
	default:
		return false
	}
}

func HandleConfigFile() {
	var c map[string]tinyini.Section
	for _, candidate := range CONFFILES {
		f, err := os.Open(candidate)
		if err != nil {
			log.Println("Configuring error: ", err)
			continue
		}
		defer f.Close()

		var errs []error
		c, errs = tinyini.Parse(f)
		if len(errs) > 0 {
			log.Println("Configuration file parse errors: ", len(errs))
			for _, err := range errs {
				log.Println(err)
			}
			continue
		}
		log.Println("Got config:", c)
		break
	}

	// Global section
	if g, ok := c[""]; ok {
		if tabszraw, ok := g["tabsize"]; ok {
			if tabsz, err := strconv.Atoi(tabszraw[0]); err != nil {
				log.Println("Invalid tabsize: ", err)
			} else {
				editorconfigs[""].TabSize = tabsz
				log.Println("Global TABSZ", tabsz)
			}
		}

		if tabspaces, ok := g["tabspaces"]; ok {
			ts := confbool(tabspaces[0])
			editorconfigs[""].TabSpaces = ts
			log.Println("TABSSPACES", ts)
		}

		// Clear ignoredirs if they are explicitly configured.
		if _, ok := g["ignoredir"]; ok {
			IGNOREDIRS = map[string]bool{}
		}
		for _, ignoredir := range g["ignoredir"] {
			log.Println("IGNOREDIR", ignoredir)
			IGNOREDIRS[ignoredir] = true
		}

		if maxfilesraw, ok := g["maxfiles"]; ok {
			if maxfiles, err := strconv.Atoi(maxfilesraw[0]); err != nil {
				log.Println("Invalid maxfiles: ", err)
			} else {
				MAXFILES = maxfiles
				log.Println("MAXFILES", MAXFILES)
			}
		}

		if worddelims, ok := g["worddelims"]; ok {
			WORD_DELIMS = worddelims[0]
			log.Println("WORDDELIMS", WORD_DELIMS)
		}

		if warnfilesizes, ok := g["warnfilesize"]; ok {
			if warnfilesize, err := strconv.ParseInt(warnfilesizes[0], 10, 64); err != nil {
				log.Println("invalid warnfilesize: ", err)
			} else {
				WARNFILESZ = warnfilesize
				log.Println("WARNFILESZ", WARNFILESZ)
			}
		}
	}

	// Handle filetype-related sections.
	for section, keyvals := range c {
		if !strings.HasPrefix(section, "filetype:") {
			continue
		}

		pattern := section[len("filetype:"):]

		if _, ok := editorconfigs[pattern]; !ok {
			nc := defaultconfig
			editorconfigs[pattern] = &nc
		}

		log.Println("pattern", pattern)
		log.Println("keyvals", keyvals)

		if savehooks, ok := keyvals["savehook"]; ok {
			sh := splitsavehook(savehooks[0])
			editorconfigs[pattern].SaveHook = sh
			log.Println(pattern, "savehook:", sh)
		}
	}
}

func getConfigFiles() (files []string) {
	if homedir, err := os.UserHomeDir(); err == nil {
		files = append(files, filepath.Join(homedir, ".ked.conf"))
	}
	if confdir, err := os.UserConfigDir(); err == nil {
		files = append(files, filepath.Join(confdir, "ked", "config"))
	}
	if len(files) == 0 {
		log.Println("Cannot determine any config file locations")
	}
	return
}

func GetEditorConfig(fpath string) (*EditorConfig, error) {
	pb := filepath.Base(fpath)
	log.Println("[GetEditorConfig] ", fpath, " -> ", pb)
	for pattern, ec := range editorconfigs {
		log.Printf("[] %q %#v\n", pattern, ec)
		matched, err := filepath.Match(pattern, pb)
		if err != nil {
			log.Printf("[GetEditorConfig, hook match] %v\n", err)
			return nil, err
		}
		if !matched {
			log.Println("[savebuffer, pattern-no-match]")
			continue
		}
		return ec, nil
	}
	return &defaultconfig, nil
}
