package main

import (
	"flag"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/buffer"
)

type editorCtx struct {
	s       tcell.Screen
	buffers []buffer.Buffer
}

func (ctx *editorCtx) initscreen() error {
	var err error
	ctx.s, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := ctx.s.Init(); err != nil {
		return err
	}
	return nil
}

func (ctx *editorCtx) mainloop() {
	quit := func() {
		ctx.s.Fini()
		log.Println("Quitting.")
		os.Exit(0)
	}
main:
	for {
		ctx.s.Show()
		ev := ctx.s.PollEvent()
		log.Printf("event: %+v\n", ev)
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			log.Printf("[resize] w=%d  h=%d\n", w, h)
			ctx.s.Sync()
			continue main
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("[quit]")
				quit()
			case ev.Key() == tcell.KeyCtrlR:
				ctx.s.Clear()
			}
		}
	}

}

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

	// Initial editor context consists of a canvas and an optional
	// list file-backed buffers.
	ctx := &editorCtx{}

	filenames := flag.Args()
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("opening buffer for file: ", filename)
		ctx.buffers = append(ctx.buffers, *buffer.NewFromFile(f))
	}

	if err := ctx.initscreen(); err != nil {
		log.Fatalln("initscreen: ", err)
	}
	ctx.mainloop()
}
