package library_test

import (
	"io/fs"
	"path"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/susji/ked/internal/testutil"
	"github.com/susji/ked/library"
)

func TestBasic(t *testing.T) {
	tfs := fstest.MapFS{
		"a":           &fstest.MapFile{Mode: fs.ModeDir},
		"a/aa":        &fstest.MapFile{Mode: fs.ModeDir},
		"b":           &fstest.MapFile{Mode: fs.ModeDir},
		"c":           &fstest.MapFile{Mode: fs.ModeDir},
		"a/a.txt":     &fstest.MapFile{Data: nil},
		"a/aa/aa.txt": &fstest.MapFile{Data: nil},
		"b/b.txt":     &fstest.MapFile{Data: nil},
		"c/c.txt":     &fstest.MapFile{Data: nil},
	}
	l := library.NewWithFS(func(_ string) fs.FS { return tfs })
	if err := l.Add("/"); err != nil {
		t.Fatal("add error: ", err)
	}

	d := "/"
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
