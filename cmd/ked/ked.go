package main

import (
	"flag"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

func main() {
	var debugfile string

	flag.StringVar(&debugfile, "debugfile", "", "File for appending debug log")
	flag.Parse()

	if len(debugfile) > 0 {
		f, err := os.OpenFile(
			debugfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
		log.Println("Opening logfile: ", debugfile)
	}

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	quit := func() {
		s.Fini()
		log.Println("Quitting.")
		os.Exit(0)
	}
main:
	for {
		s.Show()
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			continue main
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				quit()
			case ev.Key() == tcell.KeyCtrlR:
				s.Clear()
			}
		}
	}
}
