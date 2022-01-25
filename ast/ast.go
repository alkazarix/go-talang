package ast

import (
	"fmt"
	"strings"

	"github.com/alkazarix/talang/token"
)

type Node interface {
	node()
	String() string
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

// program
type Program struct {
	Statements []Stmt
}

func (p *Program) node()          {}
func (p *Program) String() string { return "program" }

// ident
type Ident struct {
	Name string
}

func (ident *Ident) node()          {}
func (ident *Ident) String() string { return ident.Name }

// literal
type Literal struct {
	Token token.Token
	Value string
}

func (*Literal) node() {}
func (*Literal) expr() {}

func (lit *Literal) String() string {
	switch lit.Token.Type {
	case token.Nil:
		return "null"
	case token.True:
		return "true"
	case token.False:
		return "false"
	case token.Number, token.String:
		return lit.Value
	}
	return ""
}

// binary expression
type BinaryExpr struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (*BinaryExpr) node() {}
func (*BinaryExpr) expr() {}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Operator.Type, e.Right)
}

// grouping expression
type GroupingExpr struct {
	Expression Expr
}

func (*GroupingExpr) node() {}
func (*GroupingExpr) expr() {}

func (e *GroupingExpr) String() string {
	return fmt.Sprintf("(%s)", e.Expression)
}

// logical expression
type LogicalExpr struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func (*LogicalExpr) node() {}
func (*LogicalExpr) expr() {}

func (e *LogicalExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Operator.Type, e.Right)
}

// unary expression
type UnaryExpr struct {
	Operator token.Token
	Right    Expr
}

func (*UnaryExpr) node() {}
func (*UnaryExpr) expr() {}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", e.Operator.Type, e.Right)
}

// var expression
type VariableExpr struct {
	Name     string
	Distance int // -1 represents global variable.
}

func (*VariableExpr) node() {}
func (*VariableExpr) expr() {}

func (e *VariableExpr) String() string {
	return e.Name
}

type ArrayExpr struct {
	Token    token.Token // the '[' token
	Elements []Expr
}

func (al *ArrayExpr) node() {}
func (al *ArrayExpr) expr() {}
func (al *ArrayExpr) String() string {
	var sb strings.Builder
	var elements []string
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	sb.WriteString(token.LeftBracket)
	sb.WriteString(strings.Join(elements, token.Comma+" "))
	sb.WriteString(token.RightBracket)

	return sb.String()
}

// assign expression
type AssignExpr struct {
	Name  string
	Value Expr
}

func (*AssignExpr) node() {}
func (*AssignExpr) expr() {}

func (e *AssignExpr) String() string {
	return fmt.Sprintf("%s = %s", e.Name, e.Value)
}

// getter expression
type GetExpr struct {
	Name token.Token
	Obj  Expr
}

func (*GetExpr) node() {}
func (*GetExpr) expr() {}

func (e *GetExpr) String() string {
	return fmt.Sprintf("%s.%s", e.Obj.String(), e.Name.Literal)
}

// setter expression
type SetExpr struct {
	Name  token.Token
	Obj   Expr
	Value Expr
}

func (*SetExpr) node() {}
func (*SetExpr) expr() {}

func (e *SetExpr) String() string {
	return fmt.Sprintf("%s.%s = %s", e.Obj.String(), e.Name.Literal, e.Value.String())
}

// this expression
type ThisExpr struct {
	Keyword token.Token
}

func (*ThisExpr) node() {}
func (*ThisExpr) expr() {}

func (e *ThisExpr) String() string {
	return fmt.Sprintf("this")
}

// super expression
type SuperExpr struct {
	Keyword token.Token
	Method  token.Token
}

func (*SuperExpr) node() {}
func (*SuperExpr) expr() {}

func (e *SuperExpr) String() string {
	return fmt.Sprintf("super")
}

// call expression

type CallExpr struct {
	Callee    Expr
	Arguments []Expr
}

func (*CallExpr) node() {}
func (*CallExpr) expr() {}

func (e *CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, arg := range e.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", e.Callee, strings.Join(args, ", "))
}

// print statement
type PrintStmt struct {
	Expression Expr
}

func (*PrintStmt) node() {}
func (*PrintStmt) stmt() {}

func (s *PrintStmt) String() string {
	var sb strings.Builder
	sb.WriteString("print ")
	sb.WriteString(s.Expression.String())
	sb.WriteRune(';')
	return sb.String()
}

// expression statement
type ExprStmt struct {
	Expression Expr
}

func (*ExprStmt) node() {}
func (*ExprStmt) stmt() {}

func (s *ExprStmt) String() string {
	return s.Expression.String() + ";"
}

// var statement
type VariableStmt struct {
	Ident       Ident
	Initializer Expr
}

func (*VariableStmt) node() {}
func (*VariableStmt) stmt() {}

func (s *VariableStmt) String() string {
	var sb strings.Builder
	sb.WriteString("let ")
	sb.WriteString(s.Ident.Name)
	sb.WriteString(" = ")
	sb.WriteString(s.Initializer.String())
	sb.WriteRune(';')
	return sb.String()
}

// block statement
type BlockStmt struct {
	Statements []Stmt
}

func (*BlockStmt) node() {}
func (*BlockStmt) stmt() {}

func (s *BlockStmt) String() string {
	var sb strings.Builder
	sb.WriteString("{ ")
	for _, stmt := range s.Statements {
		sb.WriteString(stmt.String())
	}
	sb.WriteString(" }")
	return sb.String()
}

// if statement
type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (*IfStmt) node() {}
func (*IfStmt) stmt() {}

func (s *IfStmt) String() string {
	var sb strings.Builder
	sb.WriteString("if (")
	sb.WriteString(s.Condition.String())
	sb.WriteString(") ")
	sb.WriteString(s.ThenBranch.String())
	if s.ElseBranch != nil {
		sb.WriteString(" else ")
		sb.WriteString(s.ElseBranch.String())
	}
	return sb.String()
}

// while statement
type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (*WhileStmt) node() {}
func (*WhileStmt) stmt() {}

func (s *WhileStmt) String() string {
	var sb strings.Builder
	sb.WriteString("while (")
	sb.WriteString(s.Condition.String())
	sb.WriteString(") ")
	sb.WriteString(s.Body.String())
	return sb.String()
}

// function statement
type FunctionStmt struct {
	Name          string
	Params        []*Ident
	Body          []Stmt
	IsInitializer bool
}

func (*FunctionStmt) node() {}
func (*FunctionStmt) stmt() {}

func (s *FunctionStmt) String() string {
	var sb strings.Builder
	sb.WriteString("fn ")
	sb.WriteString(s.Name)
	sb.WriteString("(")
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = p.Name
	}
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(") { ")
	for _, stmt := range s.Body {
		sb.WriteString(stmt.String())
	}
	sb.WriteString(" }")
	return sb.String()
}

// return statement
type ReturnStmt struct {
	Keyword token.Token
	Value   Expr
}

func (*ReturnStmt) node() {}
func (*ReturnStmt) stmt() {}

func (s *ReturnStmt) String() string {
	var sb strings.Builder
	sb.WriteString("return ")
	if s.Value != nil {
		sb.WriteString(s.Value.String())
	}
	sb.WriteRune(';')
	return sb.String()
}

// class statement

type ClassStmt struct {
	Name       string
	SuperClass VariableExpr
	Methods    []*FunctionStmt
}

func (*ClassStmt) node() {}
func (*ClassStmt) stmt() {}

func (s *ClassStmt) String() string {
	return "class " + s.Name
}
