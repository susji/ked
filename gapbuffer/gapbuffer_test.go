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
	b.Insert([]rune("sentence"))
	got := b.Get()
	want := []rune("this is a sentence")
	tu.Assert(t, reflect.DeepEqual(got, want), "wrong got: %q", got)
}

func TestNewFrom(t *testing.T) {
	msg := []rune("This GapBuffer has been initialized from a rune slice.")
	b := gapbuffer.NewFrom(msg)
	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msg), "unexpected got: %q", got)
}

func TestBigInsertAndCursorMove(t *testing.T) {
	b := gapbuffer.New(8)
	msg := []rune("abcdefghijklmnopqrstuvwxyz0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b.Insert(msg)
	b.SetCursor(len(msg) / 2)
	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, msg), "wrong got: %q", got)

}

func TestDeleteLots(t *testing.T) {
	msg := []rune("first second third fourth fifth sixth seventh eight ninth tenth eleventh")
	t.Logf("msg: %q", string(msg))

	left := msg[0:6]
	right := msg[6+20:]
	want := make([]rune, len(left)+len(right))
	copy(want, left)
	copy(want[len(left):], right)

	b := gapbuffer.NewFrom(msg)
	for i := 0; i < 20; i++ {
		b.SetCursor(6)
		b.Delete()
		//t.Logf("Now: %q", string(b.Get()))
	}
	got := b.Get()
	tu.Assert(t, reflect.DeepEqual(got, want), "unexpected got: %q, want: %q", string(got), string(want))
}
