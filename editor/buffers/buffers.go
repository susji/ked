package buffers

import (
	"fmt"
	"sync/atomic"

	"github.com/susji/ked/buffer"
	"github.com/susji/ked/viewport"
)

var curbufid = uint32(0)

type BufferId uint32

func newbid() uint32 {
	retid := atomic.AddUint32(&curbufid, 1)
	return retid
}

type EditorBuffers struct {
	buffers map[BufferId]*EditorBuffer
}

type EditorBuffer struct {
	Buffer                *buffer.Buffer
	Viewport              *viewport.Viewport
	bid                   uint32
	cursorline, cursorcol int
	prevsearch            string
}

func New() EditorBuffers {
	return EditorBuffers{
		buffers: map[BufferId]*EditorBuffer{},
	}
}

func (e *EditorBuffers) New(b *buffer.Buffer) BufferId {
	bid := newbid()
	neb := &EditorBuffer{
		Buffer:     b,
		Viewport:   viewport.New(b),
		cursorline: 0,
		cursorcol:  0,
		bid:        bid,
	}
	e.buffers[BufferId(bid)] = neb
	return BufferId(bid)
}

func (e *EditorBuffers) Len() int {
	return len(e.buffers)
}

func (e *EditorBuffers) Get(bid BufferId) *EditorBuffer {
	ret, ok := e.buffers[bid]
	if !ok {
		panic(fmt.Sprintf("missing bid=%d (%#v)", bid, e.buffers))
	}
	return ret
}

func (e *EditorBuffers) All() map[BufferId]*EditorBuffer {
	return e.buffers
}

func (e *EditorBuffers) Close(bid BufferId) {
	if _, ok := e.buffers[bid]; !ok {
		panic(fmt.Sprintf("missing bid=%d (%#v)", bid, e.buffers))
	}
	delete(e.buffers, bid)
}

func (eb *EditorBuffer) Update(res buffer.ActionResult) {
	eb.cursorline = res.Lineno
	eb.cursorcol = res.Col
}

func (eb *EditorBuffer) CursorLine() int {
	return eb.cursorline
}

func (eb *EditorBuffer) CursorCol() int {
	return eb.cursorcol
}

func (eb *EditorBuffer) Cursor() (lineno int, col int) {
	return eb.cursorline, eb.cursorcol
}

func (eb *EditorBuffer) SetCursor(lineno, col int) {
	eb.cursorline = lineno
	eb.cursorcol = col
}

func (eb *EditorBuffer) Id() BufferId {
	return BufferId(eb.bid)
}
