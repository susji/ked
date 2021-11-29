package gapbuffer_test

import (
	"reflect"
	"testing"

	"github.com/susji/ked/gapbuffer"
)

func assert(t *testing.T, cond bool, f string, va ...interface{}) {
	if !cond {
		t.Errorf(f, va...)
	}
}

func TestInsertGet(t *testing.T) {
	b := gapbuffer.New(16)
	msg := []rune("hello world")

	assert(t, b.Length() == 0, "should be zero length, got %d", b.Length())

	b.Insert(msg)

	assert(t, b.Length() == len(msg), "unexpected length: %d", b.Length())

	got := b.Get()
	assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)

	msg2 := []rune("yes hola ")
	msgtotal := []rune("hello yes hola world")
	b.SetCursor(6)
	b.Insert(msg2)
	got = b.Get()
	assert(t, reflect.DeepEqual(got, msgtotal), "wrong got: %q", got)
}

func TestLotsOfInserts(t *testing.T) {
	b := gapbuffer.New(8)
	msg := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	for _, r := range msg {
		b.Insert([]rune{r})
	}

	got := b.Get()
	assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)
	assert(t, len(got) == len(msg), "wrong len: %d", len(got))
}

func TestDelete(t *testing.T) {
	b := gapbuffer.New(8)
	msg := []rune("this is a very long word")
	b.Insert(msg)

	b.SetCursor(14)
	for i := 0; i < 5; i++ {
		b.Delete()
	}

	got := b.Get()
	want := []rune("this is a long word")
	assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)

	b.SetCursor(5)
	for i := 0; i < 5; i++ {
		b.Delete()
	}
	got = b.Get()
	want = []rune("is a long word")
	assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)
}
