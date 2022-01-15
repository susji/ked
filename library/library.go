// package library is responsible for maintaining a list of the files the user
// may wish to open.
package library

import (
	"fmt"
	"path/filepath"
	"sync"
)

var lib = New()

func New() *Library {
	return &Library{}
}

type Library struct {
	m    sync.RWMutex
	dirs []string
}

func (l *Library) Update() {
	l.m.Lock()
	defer l.m.Unlock()
}

func (l *Library) Add(dirpath string) {
	abs, err := filepath.Abs(dirpath)
	if err != nil {
		panic(fmt.Sprintf("library Add panic: %v", err))
	}
	l.dirs = append(l.dirs, abs)
}
