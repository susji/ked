package config

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/util"
	"github.com/susji/tinyini"
)

type EditorConfig struct {
	TabSize           int
	TabSpaces         bool
	SaveHook          []string
	HighlightPatterns []HighlightPattern
	HighlightKeywords []HighlightKeyword
}

type HighlightPattern struct {
	Priority, Left, Right int
	Pattern               string
	Style                 tcell.Style
}

type HighlightKeyword struct {
	Keyword string
	Style   tcell.Style
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

var STYLE_DEFAULT = tcell.StyleDefault
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

func parsestyle(styles string) tcell.Style {
	log.Printf("[parsestyle] %q\n", styles)
	st := STYLE_DEFAULT
	for _, style := range strings.Split(styles, ",") {
		switch style {
		case "dim":
			st = st.Dim(true)
		case "underline":
			st = st.Underline(true)
		case "bold":
			st = st.Bold(true)
		case "reverse":
			st = st.Reverse(true)
		default:
			log.Println("unrecognized style fragment: ", style)
		}
	}
	return st
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
			// Our relaxed config parsing permits errors.
			log.Println("Configuration file parse errors: ", len(errs))
			for _, err := range errs {
				log.Println(err)
			}
		}
		log.Println("Got config:", c)
		ParseConfig(f.Name(), c)
		return
	}
}

func ParseConfig(fn string, c map[string]tinyini.Section) {
	// Global section
	if g, ok := c[""]; ok {
		if tabszraw, ok := g["tabsize"]; ok {
			kv := tabszraw[0]
			if tabsz, err := strconv.Atoi(kv.Value); err != nil {
				log.Printf("%s:%d: Invalid tabsize: %v\n", fn, kv.Lineno, err)
			} else {
				editorconfigs[""].TabSize = tabsz
				log.Println("Global TABSZ", tabsz)
			}
		}

		if tabspaces, ok := g["tabspaces"]; ok {
			ts := confbool(tabspaces[0].Value)
			editorconfigs[""].TabSpaces = ts
			log.Println("TABSSPACES", ts)
		}

		// Clear ignoredirs if they are explicitly configured.
		if _, ok := g["ignoredir"]; ok {
			IGNOREDIRS = map[string]bool{}
		}
		for _, ignoredir := range g["ignoredir"] {
			log.Println("IGNOREDIR", ignoredir.Value)
			IGNOREDIRS[ignoredir.Value] = true
		}

		if maxfilesraw, ok := g["maxfiles"]; ok {
			kv := maxfilesraw[0]
			if maxfiles, err := strconv.Atoi(kv.Value); err != nil {
				log.Printf("%s:%d: Invalid maxfiles: %v\n", fn, kv.Lineno, err)
			} else {
				MAXFILES = maxfiles
				log.Println("MAXFILES", MAXFILES)
			}
		}

		if worddelims, ok := g["worddelims"]; ok {
			WORD_DELIMS = util.Unescape(worddelims[0].Value)
			log.Printf("WORDDELIMS %q\n", WORD_DELIMS)
		}

		if warnfilesizes, ok := g["warnfilesize"]; ok {
			kv := warnfilesizes[0]
			if warnfilesize, err := strconv.ParseInt(kv.Value, 10, 64); err != nil {
				log.Printf("%s:%d: invalid warnfilesize: %v\n", fn, kv.Lineno, err)
			} else {
				WARNFILESZ = warnfilesize
				log.Println("WARNFILESZ", WARNFILESZ)
			}
		}

		if savehooks, ok := g["savehook"]; ok {
			sh := splitsavehook(savehooks[0].Value)
			editorconfigs[""].SaveHook = sh
			log.Println("global savehook:", sh)
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
			sh := splitsavehook(savehooks[0].Value)
			editorconfigs[pattern].SaveHook = sh
			log.Println(pattern, "savehook:", sh)
		}

		if tabsizes, ok := keyvals["tabsize"]; ok {
			kv := tabsizes[0]
			if ts, err := strconv.Atoi(kv.Value); err != nil {
				log.Printf(
					"%s:%d: invalid tabsize for %q: %v\n",
					fn, kv.Lineno, pattern, kv)
			} else {
				editorconfigs[pattern].TabSize = ts
				log.Println(pattern, "tabsize:", ts)
			}
		}

		if tabspaces, ok := keyvals["tabspaces"]; ok {
			ts := confbool(tabspaces[0].Value)
			editorconfigs[pattern].TabSpaces = ts
			log.Println(pattern, "tabspaces:", ts)
		}

		for _, raw := range keyvals["highlight-keyword"] {
			vals := strings.SplitN(raw.Value, ":", 2)
			if len(vals) < 2 {
				log.Printf(
					"%s:%d: %s, %q: need two values for keyword highlight\n",
					fn, raw.Lineno, section, raw)
				continue
			}
			style := parsestyle(vals[0])
			keyword := vals[1]
			newkw := HighlightKeyword{
				Keyword: keyword,
				Style:   style,
			}
			log.Printf("[highlight-keyword] %#v\n", newkw)
			editorconfigs[pattern].HighlightKeywords = append(
				editorconfigs[pattern].HighlightKeywords, newkw)
		}

		for _, raw := range keyvals["highlight-pattern"] {
			vals := strings.SplitN(raw.Value, ":", 5)
			if len(vals) < 5 {
				log.Printf(
					"%s:%d: %s, %q: need five values for pattern highlight\n",
					fn, raw.Lineno, section, raw)
				continue
			}
			prio, err1 := strconv.Atoi(vals[0])
			left, err2 := strconv.Atoi(vals[1])
			right, err3 := strconv.Atoi(vals[2])
			style := parsestyle(vals[3])
			pat := vals[4]

			if err1 != nil {
				log.Printf("%s:%d: invalid priority: %v\n", fn, raw.Lineno, err1)
			}
			if err2 != nil {
				log.Printf("%s:%d: invalid left index: %v\n", fn, raw.Lineno, err2)
			}
			if err3 != nil {
				log.Printf("%s:%d: invalid right index: %v\n", fn, raw.Lineno, err3)
			}
			if err1 != nil || err2 != nil || err3 != nil {
				continue
			}

			newpat := HighlightPattern{
				Priority: prio,
				Left:     left,
				Right:    right,
				Pattern:  pat,
				Style:    style,
			}

			log.Printf("[highlight-pattern] %#v\n", newpat)
			editorconfigs[pattern].HighlightPatterns = append(
				editorconfigs[pattern].HighlightPatterns, newpat)

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

func GetEditorConfig(fpath string) *EditorConfig {
	pb := filepath.Base(fpath)
	log.Println("[GetEditorConfig] ", fpath, " -> ", pb)
	for pattern, ec := range editorconfigs {
		log.Printf("[] %q %#v\n", pattern, ec)
		matched, err := filepath.Match(pattern, pb)
		if err != nil {
			log.Printf("[GetEditorConfig, hook match] %v\n", err)
			continue
		}
		if !matched {
			continue
		}
		return ec
	}
	return &defaultconfig
}
