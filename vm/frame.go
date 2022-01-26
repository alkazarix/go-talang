package vm

import (
	"github.com/alkazarix/talang/code"
	"github.com/alkazarix/talang/valuer"
)

type Frame struct {
	fn          *valuer.CompiledFunction
	ip          int
	basePointer int
}

func NewFrame(fn *valuer.CompiledFunction, basePointer int) *Frame {
	f := &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}

	return f
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
