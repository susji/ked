package buffer

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
