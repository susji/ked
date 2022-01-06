package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/susji/ked/config"
	"github.com/susji/ked/ui/editor"
)

var (
	version   = "v0.dev"
	buildtime = "<no buildtime>"
)

func main() {
	var conffile, debugfile, ignoredirs string

	flag.StringVar(
		&conffile,
		"config",
		"",
		"Override default configuration file location")
	flag.StringVar(&debugfile, "debugfile", "", "File for appending debug log")
	flag.StringVar(
		&ignoredirs,
		"ignoredirs",
		config.GetIgnoreDirsFlat(),
		"Directories to ignore when doing buffer opens")
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "ked %s (%s)\n", version, buildtime)
		fmt.Fprintln(o, "Configuration file locations:")
		for _, fn := range config.CONFFILES {
			fmt.Fprintf(o, "  * %s\n", fn)
		}
		fmt.Fprintln(o, "")
		flag.PrintDefaults()
	}
	flag.Parse()
	config.SetConfigFile(conffile)
	config.SetIgnoreDirs(ignoredirs)

	if len(debugfile) > 0 {
		f, err := os.OpenFile(debugfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0640)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
		log.Println("Opening logfile: ", debugfile)
	} else {
		log.SetOutput(io.Discard)
	}
	log.Printf("ked %s (%s)\n", version, buildtime)

	config.HandleConfigFile()

	// Initial editor context consists of a canvas and an optional
	// list file-backed buffers.
	e := editor.New()
	filenames := flag.Args()
	for _, filename := range filenames {
		absname, err := filepath.Abs(filename)
		if err != nil {
			log.Fatalln(err)
		}
		f, err := os.Open(absname)
		if err == nil {
			log.Println("opening buffer for file: ", filename)
			e.NewBuffer(absname, f)
			f.Close()
		} else if errors.Is(err, os.ErrNotExist) {
			e.NewBuffer(absname, &bytes.Buffer{})
		} else {
			log.Fatalln(err)
		}
	}

	if err := e.Run(); err != nil {
		log.Fatalln(err)
	}
}
