package buffer

import "fmt"

const (
	MOD_LINESPLIT = iota
	MOD_INSERTRUNES
	MOD_INSERTLINE
	MOD_DELETERUNE
	MOD_DELETELINE
)

type modificationKind int

type modification struct {
	kind        modificationKind
	lineno, col int
	data        interface{}
}

func (m *modification) String() string {
	return fmt.Sprintf(
		"Modification{kind=%d, position=(%d, %d), data=%v}", m.kind, m.lineno, m.col, m.data)
}
