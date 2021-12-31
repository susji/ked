package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/susji/ked/tinyini"
)

var CONFFILES = getConfigFiles()
var TABSSPACES = false
var WARNFILESZ = int64(10_485_760)
var SAVEHOOKS = map[string][]string{}
var TABSZ = 4
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

func dosavehook(raw string) []string {
	return regexp.MustCompile(" +").Split(raw, -1)
}

func SetSaveHooks(rawsavehooks string) error {
	SAVEHOOKS = map[string][]string{}
	if len(rawsavehooks) == 0 {
		return nil
	}
	for _, rawhook := range regexp.MustCompile(" *,+ *").Split(rawsavehooks, -1) {
		parts := strings.SplitN(rawhook, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unexpected savehook given: %q", rawhook)
		}
		SAVEHOOKS[parts[0]] = dosavehook(parts[1])
	}
	return nil
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
				TABSZ = tabsz
				log.Println("TABSZ", TABSZ)
			}
		}

		if tabspaces, ok := g["tabspaces"]; ok {
			TABSSPACES = confbool(tabspaces[0])
			log.Println("TABSSPACES", TABSSPACES)
		}
	}

	// Handle filetype-related sections.
	for section, keyvals := range c {
		if !strings.HasPrefix(section, "filetype:") {
			continue
		}

		pattern := section[len("filetype:"):]
		log.Println("pattern", pattern)
		log.Println("keyvals", keyvals)

		if savehooks, ok := keyvals["savehook"]; ok {
			SAVEHOOKS[pattern] = dosavehook(savehooks[0])
			log.Println(pattern, "savehook:", savehooks[0])
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
