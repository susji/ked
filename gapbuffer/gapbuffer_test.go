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
	b := gapbuffer.New()
	msg := []rune("hello world")

	assert(t, b.Length() == 0, "should be zero length, got %d", b.Length())

	b.Insert(msg)

	assert(t, b.Length() == len(msg), "unexpected length: %d", b.Length())

	got := b.Get()
	assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)

	msg2 := []rune("yes ")
	msgtotal := []rune("hello yes world")
	b.SetCursor(6)
	b.Insert(msg2)
	got = b.Get()
	assert(t, reflect.DeepEqual(got, msgtotal), "wrong got: %q", got)

}
