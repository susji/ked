package buffer

const (
	ACT_RUNES = iota
	ACT_BACKSPACE
	ACT_LINEFEED
	ACT_DELLINECONTENT
)

type ActionKind int
type ActionFunc func(*Action) ActionResult

type Action struct {
	kind        ActionKind
	lineno, col int
	data        interface{}
}

type ActionResult struct {
	Lineno, Col int
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

func NewDelLineContent(lineno, col int) *Action {
	return &Action{
		kind:   ACT_DELLINECONTENT,
		lineno: lineno,
		col:    col,
	}
}
