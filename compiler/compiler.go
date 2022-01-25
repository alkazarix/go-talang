package compiler

import (
	"strconv"

	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/code"
	"github.com/alkazarix/talang/token"
	"github.com/alkazarix/talang/valuer"
)

var (
	Nil   = &valuer.Nil{}
	True  = &valuer.Boolean{Value: true}
	False = &valuer.Boolean{Value: false}
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []valuer.Value
}

type Compiler struct {
	instructions code.Instructions
	constants    []valuer.Value
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []valuer.Value{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExprStmt:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

	case *ast.Literal:
		value, err := literalToValue(node)
		if err != nil {
			return err
		}
		c.emit(code.OpConstant, c.addConstant(value))

	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) addConstant(value valuer.Value) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func literalToValue(literal *ast.Literal) (valuer.Value, error) {
	tok := literal.Token
	switch tok.Type {
	case token.String:
		return &valuer.String{Value: tok.Literal}, nil
	case token.Number:
		value, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			return nil, compileError(err.Error(), &tok)
		}
		return &valuer.Number{Value: value}, nil
	case token.Nil:
		return Nil, nil
	case token.True:
		return True, nil
	case token.False:
		return False, nil
	default:
		return nil, compileError("invalid token", &tok)
	}
}

func compileError(reason string, at *token.Token) error {
	err := NewCompileError(reason, at)
	return err
}
