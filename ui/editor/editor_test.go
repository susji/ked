package editor_test

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/susji/ked/ui/editor"
)

func dumpcells(t *testing.T, s tcell.SimulationScreen) {
	cells, w, h := s.GetContents()
	for i := 0; i < h; i++ {
		row := make([]rune, 0)
		for j := 0; j < w; j++ {
			row = append(row, cells[i*w+j].Runes...)
		}
		t.Log(row)
	}
}

func TestHello(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	s.SetSize(20, 3)

	e := editor.NewWithScreen(s)
	_ = e

	// Here just echo "hello" and make sure it was rendered.
	// Now, this is sort of fragile still. Clearly we want
	// to get rid of the sleeps here.
	msg := []byte("hello")
	go e.Run()
	time.Sleep(time.Second * 1)
	s.InjectKeyBytes(msg)
	time.Sleep(time.Second * 1)

	dumpcells(t, s)

	cells, _, _ := s.GetContents()
	for i, want := range append(msg, ' ') {
		rs := cells[i].Runes
		if len(rs) != 1 {
			t.Error("should have one rune but got ", len(rs))
			continue
		}
		if rs[0] != rune(want) {
			t.Errorf("wanted %c, got %c", want, rs[0])
		}
	}
}
