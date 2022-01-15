package library_test

import (
	"testing"

	"github.com/susji/ked/library"
)

func TestBasic(t *testing.T) {
	l := library.New()
	l.Update()
	l.Add(".")
}
