package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

func main() {

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	quit := func() {
		s.Fini()
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
