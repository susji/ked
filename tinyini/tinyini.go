package tinyini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
)

type Section map[string][]string

type IniError struct {
	wrapped error
	Line    int
}

var matchersection = regexp.MustCompile(`^\s*\[(.+?)\]`)
var matcherkeyval = regexp.MustCompile(`^\s*(.+?)\s*=\s*(.+?)\s*$`)
var matcherkeyvalq = regexp.MustCompile(`^\s*(.+?)\s*=\s*"((\\.|[^"\\])*)"`)
var matcherempty = regexp.MustCompile(`^\s*$`)

func (i *IniError) Error() string {
	return fmt.Sprintf("%d: %v", i.Line, i.wrapped)
}

func (i *IniError) Unwrap() error {
	return i.wrapped
}

func newError(line int, msg string) *IniError {
	return &IniError{
		wrapped: errors.New(msg),
		Line:    line,
	}
}

func Parse(r io.Reader) (map[string]Section, []error) {
	s := bufio.NewScanner(r)
	lines := []string{}
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, []error{err}
	}

	res := map[string]Section{}
	cursection := ""
	reterr := []error{}

	akv := func(key, val string) {
		if _, ok := res[cursection]; !ok {
			res[cursection] = Section{}
		}
		res[cursection][key] = append(res[cursection][key], val)
	}

	for i, line := range lines {
		if m := matcherkeyvalq.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matcherkeyval.FindStringSubmatch(line); m != nil {
			akv(m[1], m[2])
		} else if m := matchersection.FindStringSubmatch(line); m != nil {
			cursection = m[1]
		} else if m := matcherempty.FindStringIndex(line); m != nil {
			continue
		} else {
			reterr = append(reterr, newError(i+1, "not section nor key-value"))
		}
	}
	return res, reterr
}
