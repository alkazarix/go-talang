package compiler

import (
	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/code"
)

type Bytecode struct {
	Instructions code.Instructions
}

type Compiler struct {
	instructions code.Instructions
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	c.emit(code.OpNone)
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
	}
}

func (c *Compiler) emit(op code.Opcode) int {
	ins := code.Make(op)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
