package editor

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
)

type Editor struct {
	s         tcell.Screen
	buffers   []buffer.Buffer
	activebuf int
}

func New() *Editor {
	return &Editor{}
}

func (e *Editor) NewBufferFromFile(f *os.File) error {
	buf, err := buffer.NewFromFile(f)
	if err != nil {
		return err
	}
	e.buffers = append(e.buffers, *buf)
	return nil
}

func (e *Editor) initscreen() error {
	var err error
	e.s, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := e.s.Init(); err != nil {
		return err
	}
	return nil
}

func (e *Editor) Run() error {

	if err := e.initscreen(); err != nil {
		return err
	}
main:
	for {
		e.s.Show()
		ev := e.s.PollEvent()
		log.Printf("event: %+v\n", ev)
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			log.Printf("[resize] w=%d  h=%d\n", w, h)
			e.s.Sync()
			continue main
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[quit]")
				e.s.Fini()
				break main
			case ev.Key() == tcell.KeyCtrlR:
				e.s.Clear()
			}
		}
	}
	return nil
}
