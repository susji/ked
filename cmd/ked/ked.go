package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/susji/ked/editor"
)

func main() {
	var debugfile string

	flag.StringVar(&debugfile, "debugfile", "", "File for appending debug log")
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
	e := editor.New()
	filenames := flag.Args()
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("opening buffer for file: ", filename)
		e.NewBufferFromFile(f)
	}

	if err := e.Run(); err != nil {
		log.Fatalln(err)
	}
}
