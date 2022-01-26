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

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	instructions code.Instructions
	constants    []valuer.Value

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction

	scopes     []CompilationScope
	scopeIndex int

	symbolTable *SymbolTable
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	return &Compiler{
		constants:   []valuer.Value{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
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

		afterConsequencePos := len(c.currentInstructions())
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

		afterAlternativePos := len(c.currentInstructions())
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

	case *ast.FunctionStmt:
		c.enterScope()

		for _, p := range node.Params {
			c.symbolTable.Define(p.Name)
		}

		p := ast.Program{Statements: node.Body}
		err := c.Compile(&p)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		instructions := c.leaveScope()

		fmt.Printf("function instruction %s", instructions.String())

		compiledFn := &valuer.CompiledFunction{
			Instructions: instructions,
		}

		c.emit(code.OpConstant, c.addConstant(compiledFn))
		symbol := c.symbolTable.Define(node.Name)
		c.emit(code.OpSetGlobal, symbol.Index)

	case *ast.ReturnStmt:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpr:
		err := c.Compile(node.Callee)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
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
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) addConstant(value valuer.Value) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	return instructions
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
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
