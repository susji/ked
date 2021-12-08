package buffer

const (
	MOD_LINESPLIT = iota
	MOD_INSERT
	MOD_DELETE
	MOD_DELETELINE
)

type ModificationKind int

type modification struct {
	kind int
	lineno, col int
}
