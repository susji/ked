package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
		SAVEHOOKS[parts[0]] = regexp.MustCompile(" +").Split(parts[1], -1)
	}
	return nil
}

func SetConfigFile(fn string) {
	if len(fn) == 0 {
		return
	}
	CONFFILES = []string{fn}
}

func HandleConfigFile() {
	for _, candidate := range CONFFILES {
		f, err := os.Open(candidate)
		if err != nil {
			log.Println("Configuring error: ", err)
			continue
		}
		defer f.Close()

		confs, errs := tinyini.Parse(f)
		if len(errs) > 0 {
			log.Println("Configuration file parse errors: ", len(errs))
			for _, err := range errs {
				log.Println(err)
			}
			continue
		}
		log.Println("Got config:", confs)
		break
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
