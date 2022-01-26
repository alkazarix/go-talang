package compiler

import (
	"fmt"
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

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions code.Instructions
	constants    []valuer.Value

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction

	symbolTable *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []valuer.Value{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []valuer.Value) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
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

	case *ast.GroupingExpr:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

	case *ast.ExprStmt:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.LogicalExpr:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator.Literal {
		case "or":
			c.emit(code.OpOr)
		case "and":
			c.emit(code.OpAnd)
		default:
			return compileError(
				fmt.Sprintf("unknown operator %s", node.Operator.Literal),
				&node.Operator)
		}

	case *ast.BinaryExpr:
		if node.Operator.Literal == "<" || node.Operator.Literal == "<=" {

			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			switch node.Operator.Literal {
			case "<":
				c.emit(code.OpGreater)
			case "<=":
				c.emit(code.OpGreaterEqual)
			}

			return nil

		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator.Literal {
		case "+":
			c.emit(code.OpAdd)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "-":
			c.emit(code.OpSub)
		case ">":
			c.emit(code.OpGreater)
		case ">=":
			c.emit(code.OpGreaterEqual)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return compileError(
				fmt.Sprintf("unknown operator %s", node.Operator.Literal),
				&node.Operator)
		}

	case *ast.UnaryExpr:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator.Literal {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return compileError(
				fmt.Sprintf("unknown operator %s", node.Operator.Literal),
				&node.Operator)
		}

	case *ast.Literal:
		value, err := literalToValue(node)
		if err != nil {
			return err
		}
		c.emitLiteral(value)

	case *ast.IfStmt:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.ThenBranch)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
		if node.ElseBranch == nil {
			c.emit(code.OpNil)
		} else {
			err := c.Compile(node.ElseBranch)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStmt:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.VariableStmt:
		if node.Initializer != nil {
			err := c.Compile(node.Initializer)
			if err != nil {
				return err
			}
		} else {
			c.emit(code.OpNil)
		}

		symbol := c.symbolTable.Define(node.Ident.Name)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.VariableExpr:
		symbol, ok := c.symbolTable.Resolve(node.Name)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Name)
		}
		c.emit(code.OpGetGlobal, symbol.Index)

	case *ast.ArrayExpr:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
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

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) emitLiteral(value valuer.Value) {
	switch value := value.(type) {
	case *valuer.Boolean:
		if value.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *valuer.Nil:
		c.emit(code.OpNil)

	default:
		c.emit(code.OpConstant, c.addConstant(value))
	}
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

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
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
