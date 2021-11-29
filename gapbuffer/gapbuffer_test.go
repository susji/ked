package gapbuffer_test

import (
	"reflect"
	"testing"

	"github.com/susji/ked/gapbuffer"
)

func TestBasic(t *testing.T) {
	b := gapbuffer.New()
	msg := []rune("hello world")

	b.Insert(msg)
	got, n := b.Get(0, 11)
	if !reflect.DeepEqual(got, msg) {
		t.Error("wrong: ", got)
	}
	if n != 11 {
		t.Error("wrong: ", n)
	}
}
