package tinyini

import (
	"bufio"
	"io"
)

type Section map[string]string

func Parse(r io.Reader) (map[string]Section, error) {
	s := bufio.NewScanner(r)
	lines := []string{}
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}
