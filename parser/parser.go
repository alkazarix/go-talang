package parser

import (
	"fmt"
	"os"

	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func New(tokens []token.Token) *Parser {

	parser := &Parser{
		tokens:  tokens,
		current: 0,
	}
	return parser
}

func (p *Parser) Parse() (program ast.Program, err error) {

	statements := []ast.Stmt{}
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(ParseError); ok {
				program = ast.Program{}
				err = &parseErr
				p.synchronize()
			} else {
				panic(r)
			}
		}
	}()
	for !p.isAtEnd() {
		stmt := p.declaration()
		statements = append(statements, stmt)
	}
	program = ast.Program{Statements: statements}
	return program, nil
}

func (p *Parser) declaration() ast.Stmt {
	if p.match(token.Let) {
		return p.varDeclaration()
	}
	if p.match(token.Function) {
		return p.functionDeclaration()
	}
	if p.match(token.Class) {
		return p.classDeclaration()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() ast.Stmt {
	p.expect(token.Identifier, "expected `identifier` after `let`")
	indent := ast.Ident{Name: p.previous().Literal}

	var expr ast.Expr
	if p.match(token.Assign) {
		expr = p.expression()
	}
	p.expect(token.Semicolon, "expected `semicolon` after a expression")
	return &ast.VariableStmt{Ident: indent, Initializer: expr}
}

func (p *Parser) functionDeclaration() *ast.FunctionStmt {
	p.expect(token.Identifier, "expected function name")
	name := p.previous()
	p.expect(token.LeftParen, "expected '(' after function name.")
	fun := &ast.FunctionStmt{
		Name:   name.Literal,
		Params: make([]*ast.Ident, 0),
		Body:   make([]ast.Stmt, 0),
	}
	if !p.match(token.RightParen) {
		for {

			p.expect(token.Identifier, "Expect parameter name.")
			parameter := p.previous()
			if len(fun.Params) >= 255 {
				p.error("Cannot have more than 255 parameters.")
			}
			ident := &ast.Ident{Name: parameter.Literal}
			fun.Params = append(fun.Params, ident)
			if !p.match(token.Comma) {
				break
			}
		}
		p.expect(token.RightParen, "Expect ')' after parameters.")
	}
	p.expect(token.LeftBrace, "Expect '{' before function body.")
	fun.Body = p.blockStatement().Statements
	return fun
}

func (p *Parser) classDeclaration() ast.Stmt {
	p.expect(token.Identifier, "expected class name after `class`")
	name := p.previous()

	var superclass ast.VariableExpr
	if p.match(token.LessThan) {
		p.expect(token.Identifier, "expected superclass name.")
		superclass = ast.VariableExpr{
			Name: p.previous().Literal,
		}
	}

	p.expect(token.LeftBrace, "expected '{' after class name.")

	methods := make([]*ast.FunctionStmt, 0)
	for p.check(token.Identifier) {
		method := p.functionDeclaration()
		method.IsInitializer = method.Name == "init"
		methods = append(methods, method)
	}

	p.expect(token.RightBrace, "Expect '}' after class block.")

	return &ast.ClassStmt{
		Name:       name.Literal,
		Methods:    methods,
		SuperClass: superclass,
	}
}

func (p *Parser) statement() ast.Stmt {
	if p.match(token.Print) {
		return p.printStatement()
	}
	if p.match(token.LeftBrace) {
		return p.blockStatement()
	}

	if p.match(token.If) {
		return p.ifStatement()
	}

	if p.match(token.While) {
		return p.whileStatement()
	}

	if p.match(token.Return) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *Parser) printStatement() ast.Stmt {
	expr := p.expression()
	p.expect(token.Semicolon, "Expect ';' after value.")
	return &ast.PrintStmt{Expression: expr}
}

func (p *Parser) blockStatement() *ast.BlockStmt {
	statements := make([]ast.Stmt, 0)
	for !(p.check(token.RightBrace) || p.isAtEnd()) {
		statements = append(statements, p.declaration())
	}
	p.expect(token.RightBrace, "Expect '}' after block.")
	return &ast.BlockStmt{
		Statements: statements,
	}
}

func (p *Parser) ifStatement() ast.Stmt {
	p.expect(token.LeftParen, "expected `(` after if")
	expr := p.expression()
	p.expect(token.RightParen, "expected `)` after condition")

	thenBranch := p.statement()
	if p.match(token.Else) {
		elseBranch := p.statement()
		return &ast.IfStmt{Condition: expr, ThenBranch: thenBranch, ElseBranch: elseBranch}
	}
	return &ast.IfStmt{Condition: expr, ThenBranch: thenBranch}
}

func (p *Parser) whileStatement() ast.Stmt {
	p.expect(token.LeftParen, "expected `(` after while")
	expr := p.expression()
	p.expect(token.RightParen, "expected `)` after condition")
	body := p.statement()
	return &ast.WhileStmt{Condition: expr, Body: body}
}

func (p *Parser) returnStatement() ast.Stmt {
	stmt := &ast.ReturnStmt{}
	if !p.match(token.Semicolon) {
		stmt.Value = p.expression()
		p.expect(token.Semicolon, "expected ';' after return value.")
	}
	return stmt
}

func (p *Parser) expressionStatement() ast.Stmt {
	expr := p.expression()
	p.expect(token.Semicolon, "expected ';' after value.")
	return &ast.ExprStmt{Expression: expr}
}

func (p *Parser) expression() ast.Expr {
	return p.assignement()
}

func (p *Parser) assignement() ast.Expr {
	expr := p.or()
	if p.match(token.Assign) {
		value := p.assignement()
		switch e := expr.(type) {
		case *ast.VariableExpr:
			return &ast.AssignExpr{
				Name:  e.Name,
				Value: value,
			}
		case *ast.GetExpr:
			return &ast.SetExpr{
				Obj:   e.Obj,
				Name:  e.Name,
				Value: value,
			}
		default:
			p.error("invalid assignement target")
		}

	}
	return expr
}

func (p *Parser) or() ast.Expr {
	expr := p.and()
	for p.match(token.Or) {
		operator := p.previous()
		right := p.and()
		expr = &ast.LogicalExpr{Operator: operator, Left: expr, Right: right}
	}
	return expr
}

func (p *Parser) and() ast.Expr {
	expr := p.equality()
	for p.match(token.And) {
		operator := p.previous()
		right := p.equality()
		expr = &ast.LogicalExpr{Operator: operator, Left: expr, Right: right}
	}
	return expr
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparaison()
	for p.match(token.Equal, token.NotEqual) {
		operator := p.previous()
		right := p.comparaison()
		expr = &ast.BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) comparaison() ast.Expr {
	expr := p.addition()
	for p.match(token.GreaterThan, token.GreaterThanEqual, token.LessThan, token.LessThanEqual) {
		operator := p.previous()
		right := p.addition()
		expr = &ast.BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) addition() ast.Expr {
	expr := p.factor()
	for p.match(token.Plus, token.Minus) {
		operator := p.previous()
		right := p.factor()
		expr = &ast.BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) factor() ast.Expr {
	expr := p.unary()
	for p.match(token.Slash, token.Asterisk) {
		operator := p.previous()
		right := p.unary()
		expr = &ast.BinaryExpr{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}
	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(token.Bang, token.Minus) {
		operator := p.previous()
		right := p.unary()
		return &ast.UnaryExpr{
			Operator: operator,
			Right:    right,
		}
	}
	return p.call()
}

func (p *Parser) call() ast.Expr {
	expr := p.primary()
	for {
		if p.match(token.LeftParen) {
			expr = p.finishCall(expr)
		} else if p.match(token.Dot) {
			p.expect(token.Identifier, "expected property name after `.`")
			expr = &ast.GetExpr{Obj: expr, Name: p.previous()}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) finishCall(expr ast.Expr) ast.Expr {
	call := &ast.CallExpr{
		Callee:    expr,
		Arguments: make([]ast.Expr, 0),
	}
	if p.match(token.RightParen) {
		return call
	}
	for {
		arg := p.expression()
		if len(call.Arguments) >= 255 {
			p.error("function cannot have more than 255 arguments.")
		}
		call.Arguments = append(call.Arguments, arg)
		if !p.match(token.Comma) {
			break
		}
	}
	p.expect(token.RightParen, "expected  ')' after arguments.")
	return call
}

func (p *Parser) primary() (expr ast.Expr) {
	if p.match(token.True, token.False, token.Nil, token.String, token.Number) {
		tok := p.previous()
		expr = &ast.Literal{
			Token: tok,
			Value: tok.Literal,
		}

	}

	if p.match(token.This) {
		expr = &ast.ThisExpr{
			Keyword: p.previous(),
		}
	}

	if p.match(token.Super) {
		tok := p.previous()
		p.expect(token.Dot, "expected '.' after super")
		p.expect(token.Identifier, "expected method name of super")
		method := p.previous()
		expr = &ast.SuperExpr{
			Keyword: tok,
			Method:  method,
		}
	}

	if p.match(token.LeftParen) {
		inner := p.expression()
		p.expect(token.RightParen, "expected ) after expression.")
		expr = &ast.GroupingExpr{
			Expression: inner,
		}

	}

	if p.match(token.Identifier) {
		tok := p.previous()
		expr = &ast.VariableExpr{
			Name: tok.Literal,
		}
	}

	return expr
}

func (p *Parser) match(tokenTypes ...token.Type) bool {
	for _, tok := range tokenTypes {
		if p.check(tok) {
			p.avance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(tok token.Type, msg string) {
	if p.check(tok) {
		p.avance()
		return
	}
	p.error(msg)
}

func (p *Parser) check(tok token.Type) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tok
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) avance() token.Token {
	if !p.isAtEnd() {
		p.current = p.current + 1
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) synchronize() {
	p.avance()
	for !p.isAtEnd() {
		if p.previous().Type == token.Semicolon {
			return
		}
		switch p.peek().Type {
		case token.Class, token.Function, token.Let, token.If, token.While, token.Print, token.Return:
			return
		}
		p.avance()
	}
}

func (p *Parser) error(msg string) {
	s := fmt.Sprintf("%s (at line: %d, column: %d)", msg, p.peek().Position.Line, p.peek().Position.Column)
	fmt.Fprintln(os.Stderr, s)
	panic(ParseError{s})
}
