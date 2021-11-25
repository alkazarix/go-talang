package interpreter

import (
	"fmt"
	"strconv"

	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/token"
	"github.com/alkazarix/talang/valuer"
)

const (
	unknownOperatorError    = "unknown operator"
	typeMissMatchError      = "type mismatch"
	identifierNotFoundError = "identifier not found"
	notFunctionError        = "not a function"
	invalidTokenError       = "invalid token error"
)

var (
	Nil   = &valuer.Nil{}
	True  = &valuer.Boolean{Value: true}
	False = &valuer.Boolean{Value: false}
)

type Interpreter struct{}

func New() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) Evaluate(node ast.Node) valuer.Value {
	switch node := node.(type) {
	case *ast.Program:
		return i.evalProgram(node)
	case *ast.ExprStmt:
		return i.Evaluate(node.Expression)
	case *ast.Literal:
		return i.evalLiteral(node)
	case *ast.UnaryExpr:
		return i.evalUnary(node)
	case *ast.BinaryExpr:
		return i.evalBinary(node)
	case *ast.LogicalExpr:
		return i.evalLogical(node)
	}

	return nil
}

func (i *Interpreter) evalProgram(program *ast.Program) valuer.Value {
	var result valuer.Value
	for _, statement := range program.Statements {
		result = i.Evaluate(statement)
		if result.Type() == valuer.ErrorType {
			return result
		}
	}
	return result
}

func (i *Interpreter) evalLiteral(literal *ast.Literal) valuer.Value {
	tok := literal.Token
	switch tok.Type {
	case token.String:
		return &valuer.String{Value: tok.Literal}
	case token.Number:
		value, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			return &valuer.Error{Error: err}
		}
		return &valuer.Number{Value: value}
	case token.Nil:
		return Nil
	case token.True:
		return True
	case token.False:
		return False
	default:
		return &valuer.Error{Error: NewRuntimeError(invalidTokenError, tok)}
	}
}

func (i *Interpreter) evalUnary(unary *ast.UnaryExpr) valuer.Value {

	operator := unary.Operator

	right := i.Evaluate(unary.Right)
	if isError(right) {
		return right
	}

	switch operator.Type {
	case token.Bang:
		return i.evalBangOperator(right)
	case token.Minus:
		if right.Type() != valuer.NumberType {
			msg := fmt.Sprintf("%s: -%s", unknownOperatorError, right.Type())
			return newError(msg, operator)
		}

		value := right.(*valuer.Number).Value
		return &valuer.Number{Value: -value}
	default:
		msg := fmt.Sprintf("%s: %s%s", unknownOperatorError, operator.Type, right.Inspect())
		return newError(msg, operator)
	}
}

func (i *Interpreter) evalBangOperator(right valuer.Value) valuer.Value {
	switch right {
	case True:
		return False
	case False:
		return True
	case Nil:
		return True
	default:
		return False
	}
}

func (i *Interpreter) evalBinary(binary *ast.BinaryExpr) valuer.Value {

	left := i.Evaluate(binary.Left)
	if isError(left) {
		return left
	}
	rigth := i.Evaluate(binary.Right)
	if isError(rigth) {
		return rigth
	}

	if left.Type() == valuer.NumberType && rigth.Type() == valuer.NumberType {
		return i.evalBinaryNumber(binary.Operator, left, rigth)
	}

	if left.Type() == valuer.StringType && rigth.Type() == valuer.StringType {
		return i.evalBinaryString(binary.Operator, left, rigth)
	}

	msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, binary.Operator.Type, left.Inspect(), rigth.Inspect())
	return newError(msg, binary.Operator)
}

func (i *Interpreter) evalLogical(logical *ast.LogicalExpr) valuer.Value {
	left := i.Evaluate(logical.Left)
	if isError(left) {
		return left
	}
	right := i.Evaluate(logical.Right)
	if isError(right) {
		return right
	}

	leftValue := isTruthy(left)
	rightValue := isTruthy(right)

	switch logical.Operator.Type {
	case token.And:
		return toBoolanValue(leftValue && rightValue)
	case token.Or:
		return toBoolanValue(leftValue || rightValue)
	default:
		msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, logical.Operator.Type, left.Inspect(), right.Inspect())
		return newError(msg, logical.Operator)

	}
}

func (i *Interpreter) evalBinaryNumber(operator token.Token, left, right valuer.Value) valuer.Value {
	leftValue := left.(*valuer.Number).Value
	rightValue := right.(*valuer.Number).Value

	switch operator.Type {
	case token.Plus:
		return &valuer.Number{Value: leftValue + rightValue}
	case token.Minus:
		return &valuer.Number{Value: leftValue - rightValue}
	case token.Asterisk:
		return &valuer.Number{Value: leftValue * rightValue}
	case token.Slash:
		return &valuer.Number{Value: leftValue / rightValue}
	case token.LessThan:
		return toBoolanValue(leftValue < rightValue)
	case token.GreaterThan:
		return toBoolanValue(leftValue > rightValue)
	case token.Equal:
		return toBoolanValue(leftValue == rightValue)
	case token.NotEqual:
		return toBoolanValue(leftValue != rightValue)
	default:
		msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, left.Type(), operator.Type, right.Type())

		return newError(msg, operator)
	}
}

func (i *Interpreter) evalBinaryString(operator token.Token, left, right valuer.Value) valuer.Value {
	leftValue := left.(*valuer.String).Value
	rightValue := right.(*valuer.String).Value

	if operator.Type == token.Plus {
		return &valuer.String{Value: leftValue + rightValue}
	}
	msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, left.Type(), operator.Type, right.Type())
	return newError(msg, operator)

}

func isTruthy(obj valuer.Value) bool {
	switch obj {
	case Nil:
		return false
	case True:
		return true
	case False:
		return false
	default:
		return true
	}
}

func toBoolanValue(input bool) *valuer.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func isError(value valuer.Value) bool {
	if value.Type() == valuer.ErrorType {
		return true
	}
	return false
}

func newError(reason string, at token.Token) *valuer.Error {
	err := NewRuntimeError(reason, at)
	return &valuer.Error{Error: err}
}
