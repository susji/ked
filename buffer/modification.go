package buffer

import "fmt"

const (
	MOD_INSERTRUNES = iota
	MOD_LINEFEED
	MOD_DELETERUNES
	MOD_DELETELINE
	MOD_MOVERUNES
)

var kindnames = map[modificationKind]string{
	MOD_INSERTRUNES: "MOD_INSERTRUNES",
	MOD_LINEFEED:    "MOD_LINEFEED",
	MOD_DELETERUNES: "MOD_DELETERUNES",
	MOD_DELETELINE:  "MOD_DELETELINE",
	MOD_MOVERUNES:   "MOD_MOVERUNES",
}

type modificationKind int

type modification struct {
	kind        modificationKind
	lineno, col int
	data        interface{}
}

func (m *modification) String() string {
	return fmt.Sprintf(
		"Modification{kind=%s, position=(%d, %d), data=%v}",
		kindnames[m.kind], m.lineno, m.col, m.data)
}
