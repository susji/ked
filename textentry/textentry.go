package textentry

import (
	"errors"
)

var (
	Cancel   = errors.New("text entry cancelled")
	Done     = errors.New("text entry done")
	Overflow = errors.New("text entry overflowing")
	Delete   = errors.New("delete last character from text entry")
)

type RenderFunc func(column int, r rune) error
type RuneFunc func() (rune, error)

type TextEntry struct {
	defval, prompt string
	maxlen         int
}

func New(defval, prompt string, maxlen int) *TextEntry {
	return &TextEntry{
		defval: defval,
		prompt: prompt,
		maxlen: maxlen,
	}
}

func (te *TextEntry) Ask(uf RuneFunc, rf RenderFunc) ([]rune, error) {
	for i, r := range te.prompt {
		rf(i, r)
	}
	for i, r := range te.defval {
		rf(i+len(te.prompt), r)
	}
	answer := []rune(te.defval)
	var r rune
	var runeerr error
	for {
		r, runeerr = uf()
		if runeerr != nil && (errors.Is(Cancel, runeerr) || errors.Is(Done, runeerr)) {
			break
		}

		if runeerr != nil && errors.Is(Delete, runeerr) && len(answer) > 0 {
			answer = answer[:len(answer)-1]
			rf(len(te.prompt)+len(answer), ' ')
		} else {
			answer = append(answer, r)
			rf(len(te.prompt)+len(answer)-1, r)

		}
		if len(answer) >= te.maxlen {
			break
		}
	}
	if runeerr != nil && !errors.Is(runeerr, Done) {
		return nil, runeerr
	}
	return answer, nil
}
