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
