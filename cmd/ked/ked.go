package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/susji/ked/editor"
)

func main() {
	var debugfile, savehook string

	flag.StringVar(&debugfile, "debugfile", "", "File for appending debug log")
	flag.StringVar(&savehook, "savehook", "",
		"Command to run when a file is saved. __ABSPATH__ is expanded to filepath.")
	flag.Parse()

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
