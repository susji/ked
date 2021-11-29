package gapbuffer_test

import (
	"reflect"
	"testing"

	"github.com/susji/ked/gapbuffer"
	tu "github.com/susji/ked/internal/testutil"
)

func TestInsertGet(t *testing.T) {
	b := gapbuffer.New(16)
	msg := []rune("hello world")

	tu.Assert(t, b.Length() == 0, "should be zero length, got %d", b.Length())

	b.Insert(msg)

	tu.Assert(t, b.Length() == len(msg), "unexpected length: %d", b.Length())

	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)

	msg2 := []rune("yes hola ")
	msgtotal := []rune("hello yes hola world")
	b.SetCursor(6)
	b.Insert(msg2)
	got = b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msgtotal), "wrong got: %q", got)
}

func TestLotsOfInserts(t *testing.T) {
	b := gapbuffer.New(8)
	msg := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	for _, r := range msg {
		b.Insert([]rune{r})
	}

	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)
	tu.Assert(t, len(got) == len(msg), "wrong len: %d", len(got))
}

func TestDelete(t *testing.T) {
	b := gapbuffer.New(8)
	msg := []rune("this is a very long phrase")
	b.Insert(msg)

	b.SetCursor(14)
	for i := 0; i < len(" very"); i++ {
		b.Delete()
	}

	got := b.Get()
	want := []rune("this is a long phrase")
	tu.Assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)

	b.SetCursor(5)
	for i := 0; i < len("this "); i++ {
		b.Delete()
	}
	got = b.Get()
	want = []rune("is a long phrase")
	tu.Assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)
}

func TestCursorExtremes(t *testing.T) {
	b := gapbuffer.New(0)

	msg := []rune("this is a phrase")
	b.Insert(msg)
	b.SetCursor(0)
	b.SetCursor(len(msg))
	b.SetCursor(0)

	b.SetCursor(len(msg))
	for i := 0; i < len("phrase"); i++ {
		b.Delete()
	}
	print("ZZZ", b.Get())
	b.Insert([]rune("sentence"))
	got := b.Get()
	want := []rune("this is a sentence")
	tu.Assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)
}

func TestNewFrom(t *testing.T) {
	msg := []rune("This GapBuffer has been initialized from a rune slice.")
	b := gapbuffer.NewFrom([]rune(msg))
	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msg), "unexpected got: %q", got)
}
