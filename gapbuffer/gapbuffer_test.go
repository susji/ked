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

func TestBasic(t *testing.T) {
	b := gapbuffer.New()
	msg := []rune("hello world")

	assert(t, b.Length() == 0, "should be zero length, got %d", b.Length())

	b.Insert(msg)

	assert(t, b.Length() == len(msg), "unexpected length: %d", b.Length())

	got, n := b.Get(0, 11)
	assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)
	assert(t, n == 11, "wrong n: %d", n)

	got, n = b.Get(5, 5)
	assert(t, reflect.DeepEqual(got, msg[5:5+5]), "wrong got: %q", got)
	assert(t, n == 5, "wrong n: %d", n)

	got, n = b.Get(0, 10000)
	assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)
	assert(t, n == 11, "wrong n: %d", n)

}
