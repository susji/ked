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
	"github.com/susji/ked/editor"
)

var (
	version   = "v0.dev"
	buildtime = "<no buildtime>"
)

func main() {
	var debugfile, savehook, ignoredirs string

	flag.StringVar(&debugfile, "debugfile", "", "File for appending debug log")
	flag.StringVar(&savehook, "savehook", "",
		"Command to run when a file is saved. __ABSPATH__ is expanded to filepath.")
	flag.IntVar(&config.TABSZ, "tabsize", config.TABSZ, "Tab size")
	flag.StringVar(
		&ignoredirs,
		"ignoredirs",
		config.GetIgnoreDirsFlat(),
		"Directories to ignore when doing buffer opens")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "ked %s (%s)\n", version, buildtime)
		flag.PrintDefaults()
	}
	flag.Parse()
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

	// Initial editor context consists of a canvas and an optional
	// list file-backed buffers.
	e := editor.New().SaveHook(savehook)
	filenames := flag.Args()
	for _, filename := range filenames {
		absname, err := filepath.Abs(filename)
		if err != nil {
			log.Fatalln(err)
		}
		f, err := os.Open(absname)
		if err == nil {
			log.Println("opening buffer for file: ", filename)
			e.NewBufferFromFile(f)
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
