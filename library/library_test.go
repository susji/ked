package library_test

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/susji/ked/internal/testutil"
	"github.com/susji/ked/library"
)

func TestBasic(t *testing.T) {
	l := library.New()
	if err := l.Add("testdata"); err != nil {
		t.Fatal("add error: ", err)
	}

	wd, _ := os.Getwd()
	d := path.Join(wd, "testdata")
	want := map[string]bool{
		path.Join(d, "a", "a.txt"):        true,
		path.Join(d, "a", "aa", "aa.txt"): true,
		path.Join(d, "b", "b.txt"):        true,
		path.Join(d, "c", "c.txt"):        true,
	}

	got := map[string]bool{}
	l.Walk(func(filename string) error {
		t.Logf("got filename: %s\n", filename)
		got[filename] = true
		return nil
	})

	testutil.Assert(
		t,
		reflect.DeepEqual(want, got),
		"unexpected got: %#v, want %#v",
		got,
		want)
}
