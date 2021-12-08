package buffer

const (
	ACT_RUNES = iota
	ACT_BACKSPACE
	ACT_LINEFEED
	ACT_DELLINE
)

type Action struct {
	kind        int
	lineno, col int
	data        interface{}
}

func NewInsert(lineno, col int, rs []rune) *Action {
	return &Action{
		kind:   ACT_RUNES,
		lineno: lineno,
		col:    col,
		data:   rs,
	}
}

func NewBackspace(lineno, col int) *Action {
	return &Action{
		kind:   ACT_BACKSPACE,
		lineno: lineno,
		col:    col,
	}
}

func NewLinefeed(lineno, col int) *Action {
	return &Action{
		kind:   ACT_LINEFEED,
		lineno: lineno,
		col:    col,
	}
}

func NewDelLine(lineno int) *Action {
	return &Action{
		kind:   ACT_DELLINE,
		lineno: lineno,
	}
}
