package interpreter

import (
	"errors"
	"fmt"
	"reflect"
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
	invalidTokenError       = "invalid token"
	requiredInstanceError   = "required instance"
	propertyNotFoundError   = "undefined propterty"
)

var (
	Nil   = &valuer.Nil{}
	True  = &valuer.Boolean{Value: true}
	False = &valuer.Boolean{Value: false}
)

type Interpreter struct {
	env *valuer.Environment
}

func New() *Interpreter {
	env := valuer.NewEnvironment()

	env.Define("clock", &valuer.Clock{})
	env.Define("at", &valuer.At{})
	env.Define("len", &valuer.Len{})
	env.Define("push", &valuer.Push{})
	env.Define("rest", &valuer.Rest{})

	return &Interpreter{
		env: env,
	}
}

func (i *Interpreter) Evaluate(node ast.Node) (value valuer.Value, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				panic(r)
			}
		}
	}()

	value = i.eval(node)
	return value, err
}

func (i *Interpreter) eval(node ast.Node) valuer.Value {
	switch node := node.(type) {
	case *ast.Program:
		return i.evalProgram(node)
	case *ast.ExprStmt:
		return i.eval(node.Expression)
	case *ast.Literal:
		return i.evalLiteral(node)
	case *ast.ArrayExpr:
		return i.evalArray(node)
	case *ast.UnaryExpr:
		return i.evalUnary(node)
	case *ast.BinaryExpr:
		return i.evalBinary(node)
	case *ast.LogicalExpr:
		return i.evalLogical(node)
	case *ast.GroupingExpr:
		return i.eval(node.Expression)
	case *ast.VariableStmt:
		return i.evalVariableStmt(node)
	case *ast.AssignExpr:
		return i.evalAssignExpr(node)
	case *ast.VariableExpr:
		return i.evalVariableExpr(node)
	case *ast.CallExpr:
		return i.evalCallExpr(node)
	case *ast.GetExpr:
		return i.evalGetExpr(node)
	case *ast.SetExpr:
		return i.evalSetExpr(node)
	case *ast.ThisExpr:
		return i.evalThisExpr(node)
	case *ast.BlockStmt:
		return i.evalBlockStmt(node)
	case *ast.ReturnStmt:
		return i.evalReturnStmt(node)
	case *ast.IfStmt:
		return i.evalIfStmt(node)
	case *ast.WhileStmt:
		return i.evalWhileStmt(node)
	case *ast.FunctionStmt:
		return i.evalFunctionStmt(node)
	case *ast.ClassStmt:
		return i.evalClassStmt(node)
	case *ast.PrintStmt:
		return i.evalPrintStmt(node)
	default:
		panic(fmt.Sprintf("unknown ast type %#v.", node))
	}
}

func (i *Interpreter) evalProgram(program *ast.Program) valuer.Value {
	var result valuer.Value
	for _, statement := range program.Statements {
		result = i.eval(statement)

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
			runtimeError(err.Error(), &tok)
			return nil
		}
		return &valuer.Number{Value: value}
	case token.Nil:
		return Nil
	case token.True:
		return True
	case token.False:
		return False
	default:
		runtimeError(invalidTokenError, &tok)
		return nil
	}
}

func (i *Interpreter) evalArray(array *ast.ArrayExpr) valuer.Value {
	elements := []valuer.Value{}
	for _, expr := range array.Elements {
		element := i.eval(expr)
		elements = append(elements, element)
	}
	return &valuer.Array{Elements: elements}
}

func (i *Interpreter) evalUnary(unary *ast.UnaryExpr) valuer.Value {

	operator := unary.Operator
	right := i.eval(unary.Right)

	switch operator.Type {
	case token.Bang:
		return i.evalBangOperator(right)
	case token.Minus:
		if right.Type() != valuer.NumberType {
			msg := fmt.Sprintf("%s: -%s", unknownOperatorError, right.Type())
			runtimeError(msg, &operator)
			return nil
		}

		value := right.(*valuer.Number).Value
		return &valuer.Number{Value: -value}
	default:
		msg := fmt.Sprintf("%s: %s%s", unknownOperatorError, operator.Type, right.Inspect())
		runtimeError(msg, &operator)
		return nil
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

	left := i.eval(binary.Left)
	rigth := i.eval(binary.Right)

	operator := binary.Operator
	switch {
	case left.Type() == valuer.NumberType && rigth.Type() == valuer.NumberType:
		return i.evalBinaryNumber(operator, left, rigth)
	case left.Type() == valuer.StringType && rigth.Type() == valuer.StringType:
		return i.evalBinaryString(operator, left, rigth)
	case operator.Type == token.Equal:
		return toBoolanValue(left == rigth)
	case operator.Type == token.NotEqual:
		return toBoolanValue(left != rigth)
	case left.Type() != rigth.Type():
		msg := fmt.Sprintf("%s: %s %s %s", typeMissMatchError, left.Type(), operator.Literal, rigth.Type())
		runtimeError(msg, &operator)
	default:
		msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, left.Type(), operator.Literal, rigth.Type())
		runtimeError(msg, &operator)
	}
	return nil
}

func (i *Interpreter) evalLogical(logical *ast.LogicalExpr) valuer.Value {
	left := i.eval(logical.Left)
	right := i.eval(logical.Right)

	leftValue := isTruthy(left)
	rightValue := isTruthy(right)

	switch logical.Operator.Type {
	case token.And:
		return toBoolanValue(leftValue && rightValue)
	case token.Or:
		return toBoolanValue(leftValue || rightValue)
	default:
		msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, logical.Operator.Type, left.Inspect(), right.Inspect())
		runtimeError(msg, &logical.Operator)
		return nil
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
		runtimeError(msg, &operator)
		return nil
	}
}

func (i *Interpreter) evalBinaryString(operator token.Token, left, right valuer.Value) valuer.Value {
	leftValue := left.(*valuer.String).Value
	rightValue := right.(*valuer.String).Value

	if operator.Type != token.Plus {
		msg := fmt.Sprintf("%s: %s %s %s", unknownOperatorError, left.Type(), operator.Type, right.Type())
		runtimeError(msg, &operator)
	}
	return &valuer.String{Value: leftValue + rightValue}
}

func (i *Interpreter) evalVariableStmt(stmt *ast.VariableStmt) valuer.Value {
	name := stmt.Ident.Name
	var v valuer.Value
	if stmt.Initializer != nil {
		v = i.eval(stmt.Initializer)
	} else {
		v = Nil
	}
	i.env.Define(name, v)

	return Nil
}

func (i *Interpreter) evalVariableExpr(expr *ast.VariableExpr) valuer.Value {
	if v, ok := i.env.Get(expr.Name); ok {
		return v
	}

	msg := fmt.Sprintf("%s: %s", identifierNotFoundError, expr.Name)
	runtimeError(msg, nil)
	return nil
}

func (i *Interpreter) evalAssignExpr(expr *ast.AssignExpr) valuer.Value {
	v := i.eval(expr.Value)
	if ok := i.env.Assign(expr.Name, v); ok {
		return v
	}
	msg := fmt.Sprintf("%s: %s", identifierNotFoundError, expr.Name)
	runtimeError(msg, nil)
	return nil
}

func (i *Interpreter) evalCallExpr(expr *ast.CallExpr) valuer.Value {
	callee := i.eval(expr.Callee)

	fmt.Printf("called %s\n", callee.Inspect())
	callableValue, ok := callee.(valuer.Callable)
	if !ok {
		msg := fmt.Sprintf("%s: %s", notFunctionError, expr.Callee.String())
		runtimeError(msg, nil)
	}
	if l, l1 := callableValue.Arity(), len(expr.Arguments); l != l1 {
		msg := fmt.Sprintf("Expected %d arguments but got %d", l, l1)
		runtimeError(msg, nil)
	}

	if callee.Type() == valuer.BuiltinType {
		fn := reflect.ValueOf(callee).MethodByName("Fn")
		retv := fn.Call([]reflect.Value{})
		buildIn := retv[0].Interface().(valuer.BuiltinFunction)
		params := []valuer.Value{}
		for _, arg := range expr.Arguments {
			params = append(params, i.eval(arg))
		}

		returnValue, err := buildIn(params...)
		if err != nil {
			runtimeError(err.Error(), nil)
		}

		return returnValue
	}

	switch node := callee.(type) {
	default:
		panic("invalid type")
	case *valuer.Function:
		return i.callFunction(node, expr.Arguments)
	case *valuer.Klass:
		return i.ctorInstance(node, expr.Arguments)
	}
}

func (i *Interpreter) evalSetExpr(expr *ast.SetExpr) valuer.Value {
	obj := i.eval(expr.Obj)
	instance, ok := obj.(*valuer.Instance)
	if !ok {
		msg := fmt.Sprintf("%s: got %s", requiredInstanceError, obj.Type())
		runtimeError(msg, &expr.Name)
	}
	v := i.eval(expr.Value)
	instance.Set(expr.Name.Literal, v)
	return v
}

func (i *Interpreter) evalGetExpr(expr *ast.GetExpr) valuer.Value {
	obj := i.eval(expr.Obj)
	instance, ok := obj.(*valuer.Instance)
	if !ok {
		msg := fmt.Sprintf("%s: got %s", requiredInstanceError, obj.Type())
		runtimeError(msg, &expr.Name)
	}
	if property, ok := instance.Get(expr.Name.Literal); ok {
		return property
	}
	msg := fmt.Sprintf("%s: want %s", propertyNotFoundError, expr.Name.Literal)
	runtimeError(msg, &expr.Name)
	return nil

}

func (i *Interpreter) evalThisExpr(expr *ast.ThisExpr) valuer.Value {
	if v, ok := i.env.Get("this"); ok {
		return v
	}
	msg := fmt.Sprintf("could not use `this` outside a class")
	runtimeError(msg, &expr.Keyword)
	return nil
}

func (i *Interpreter) callFunction(fn *valuer.Function, args []ast.Expr) valuer.Value {
	environment := fn.Closure
	environment = valuer.NewEnclosing(fn.Closure)
	for index, param := range fn.Params {
		environment.Define(param.Name, i.eval(args[index]))
	}
	v := i.executeBlock(fn.Body, environment)
	if returnValue, ok := v.(*valuer.Return); ok {
		return returnValue.Value
	}
	return Nil
}

func (i *Interpreter) ctorInstance(klass *valuer.Klass, args []ast.Expr) valuer.Value {
	instance := &valuer.Instance{Klass: klass}
	initializer := klass.FindMethod("init")
	if initializer != nil {
		i.callFunction(initializer.Bind(instance), args)
	}
	return instance
}

func (i *Interpreter) evalBlockStmt(stmt *ast.BlockStmt) valuer.Value {
	enclosingEnv := valuer.NewEnclosing(i.env)
	return i.executeBlock(stmt.Statements, enclosingEnv)

}

func (i *Interpreter) evalIfStmt(stmt *ast.IfStmt) valuer.Value {
	condition := i.eval(stmt.Condition)
	if isTruthy(condition) {
		return i.eval(stmt.ThenBranch)
	}
	if stmt.ElseBranch == nil {
		return Nil
	}
	return i.eval(stmt.ElseBranch)
}

func (i *Interpreter) evalWhileStmt(stmt *ast.WhileStmt) valuer.Value {
	for isTruthy(i.eval(stmt.Condition)) {
		result := i.eval(stmt.Body)
		if result != nil {
			if rt := result.Type(); rt == valuer.ReturnType {
				return result
			}
		}
	}
	return Nil
}

func (i *Interpreter) evalFunctionStmt(stmt *ast.FunctionStmt) valuer.Value {
	v := valuer.Function{
		Name:          stmt.Name,
		Params:        stmt.Params,
		Body:          stmt.Body,
		IsInitializer: stmt.IsInitializer,
		Closure:       i.env,
	}

	i.env.Define(stmt.Name, &v)
	return Nil
}

func (i *Interpreter) evalClassStmt(stmt *ast.ClassStmt) valuer.Value {
	methods := make(map[string]*valuer.Function)
	for _, method := range stmt.Methods {
		fn := &valuer.Function{
			Name:          method.Name,
			Params:        method.Params,
			Body:          method.Body,
			Closure:       i.env,
			IsInitializer: method.IsInitializer,
		}
		methods[method.Name] = fn
	}
	klass := valuer.Klass{
		Name:    stmt.Name,
		Methods: methods,
	}
	i.env.Define(klass.Name, &klass)
	return Nil
}

func (i *Interpreter) evalPrintStmt(stmt *ast.PrintStmt) valuer.Value {
	v := i.eval(stmt.Expression)
	fmt.Println(v.Inspect())
	return Nil
}

func (i *Interpreter) executeBlock(stmts []ast.Stmt, env *valuer.Environment) valuer.Value {
	previous := i.env
	i.env = env

	var result valuer.Value = Nil
	defer func() {
		i.env = previous
	}()

	for _, stmt := range stmts {
		result = i.eval(stmt)
		if result != nil && result.Type() == valuer.ReturnType {
			return result
		}
	}
	return result
}

func (i *Interpreter) evalReturnStmt(stmt *ast.ReturnStmt) valuer.Value {
	var v valuer.Value = Nil
	if stmt.Value != nil {
		v = i.eval(stmt.Value)
	}
	return &valuer.Return{Value: v}
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

func runtimeError(reason string, at *token.Token) {
	err := NewRuntimeError(reason, at)
	panic(err)
}
