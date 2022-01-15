// package library is responsible for maintaining a list of the files the user
// may wish to open.
package library

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/susji/ked/config"
)

var (
	ErrorMaxFiles = errors.New("maximum amount of files encountered")
	lib           = New()
)

func New() *Library {
	return &Library{
		libfs:     os.DirFS,
		maxfiles:  config.MAXFILES,
		filepaths: map[string]struct{}{},
	}
}

func NewWithFS(libfs func(string) fs.FS) *Library {
	l := New()
	l.libfs = libfs
	return l
}

type Library struct {
	libfs     func(string) fs.FS
	maxfiles  int
	m         sync.RWMutex
	filepaths map[string]struct{}
}

type Walker func(filepath string) error

func (l *Library) Reset() {
	l.filepaths = map[string]struct{}{}
}

func (l *Library) update(absdir string) error {
	l.m.Lock()
	defer l.m.Unlock()
	return fs.WalkDir(l.libfs(absdir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("library update failed: %v\n", err)
			return err
		}
		if d.IsDir() {
			if _, ok := config.IGNOREDIRS[filepath.Base(path)]; ok {
				return fs.SkipDir
			}
			return nil
		}
		if len(l.filepaths) >= config.MAXFILES {
			return ErrorMaxFiles
		}
		l.filepaths[filepath.Join(absdir, path)] = struct{}{}
		return nil
	})
}

func (l *Library) Add(dirpath string) error {
	abs, err := filepath.Abs(dirpath)
	if err != nil {
		panic(fmt.Sprintf("library Add panic: %v", err))
	}
	return l.update(abs)
}

func (l *Library) Walk(fn Walker) error {
	l.m.RLock()
	for filepath, _ := range l.filepaths {
		fn(filepath)
	}
	defer l.m.RUnlock()
	return nil
}
