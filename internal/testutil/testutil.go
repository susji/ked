package testutil

import (
	"path"
	"runtime"
	"testing"
)

func Assert(t *testing.T, cond bool, f string, va ...interface{}) {
	if !cond {
		_, file, lineno, _ := runtime.Caller(1)
		args := []interface{}{}
		args = append(args, path.Base(file))
		args = append(args, lineno)
		args = append(args, va...)
		t.Errorf("<%s:%d> "+f, args...)
	}
}
