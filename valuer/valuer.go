package valuer

import (
	"fmt"
	"strings"

	"github.com/alkazarix/talang/ast"
)

type ValueType string

const (
	NumberType   = "Number"
	BooleanType  = "Boolean"
	NilType      = "Nil"
	StringType   = "String"
	ArrayType    = "Array"
	ReturnType   = "Return"
	KlassType    = "Klass"
	InstanceType = "Instance"
	FunctionType = "Function"
	BuiltinType  = "Builtin"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

type Callable interface {
	Arity() int
	call()
}

// number value
type Number struct {
	Value float64
}

func (n *Number) Type() ValueType { return NumberType }
func (n *Number) Inspect() string {
	if n.Value == float64(int(n.Value)) {
		return fmt.Sprintf("%d", int(n.Value))
	}
	return fmt.Sprintf("%f", n.Value)
}

// boolean value
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ValueType { return BooleanType }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// null value
type Nil struct{}

func (n *Nil) Type() ValueType { return NilType }
func (n *Nil) Inspect() string { return "nil" }

// string value
type String struct {
	Value string
}

func (s *String) Type() ValueType { return StringType }
func (s *String) Inspect() string { return s.Value }

// return value
type Return struct {
	Value Value
}

func (r *Return) Type() ValueType { return ReturnType }
func (r *Return) Inspect() string { return r.Value.Inspect() }

// array value
type Array struct {
	Elements []Value
}

func (a *Array) Type() ValueType { return ArrayType }
func (a *Array) Inspect() string {
	var sb strings.Builder

	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	sb.WriteString("[")
	sb.WriteString(strings.Join(elements, ", "))
	sb.WriteString("]")

	return sb.String()
}

// function value
type Function struct {
	Name          string
	Params        []*ast.Ident
	Body          []ast.Stmt
	Closure       *Environment
	IsInitializer bool
}

func (*Function) Type() ValueType { return FunctionType }

func (*Function) call() {}

func (fn *Function) Inspect() string {
	return "<fn " + fn.Name + ">"
}

func (fn *Function) Arity() int {
	return len(fn.Params)
}

func (fn *Function) Bind(i *Instance) *Function {
	env := NewEnclosing(fn.Closure)
	env.Define("this", i)
	return &Function{
		Name:    fn.Name,
		Params:  fn.Params,
		Body:    fn.Body,
		Closure: env,
	}
}

type Klass struct {
	Name    string
	Methods map[string]*Function
}

func (*Klass) Type() ValueType { return KlassType }

func (*Klass) call() {}

func (k *Klass) Arity() int {
	initializer := k.FindMethod("init")
	if initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (k *Klass) Inspect() string {
	return "class " + k.Name
}

func (k *Klass) FindMethod(key string) *Function {
	if method, ok := k.Methods[key]; ok {
		return method
	}
	return nil
}

type Instance struct {
	Klass  *Klass
	Fields map[string]Value
}

func (*Instance) Type() ValueType { return InstanceType }

func (i *Instance) Inspect() string {
	return i.Klass.Name + " instance"
}

func (i *Instance) Get(key string) (Value, bool) {
	if v, ok := i.Fields[key]; ok {
		return v, ok
	}
	if method := i.Klass.FindMethod(key); method != nil {
		return method.Bind(i), true
	}
	return nil, false
}

func (i *Instance) Set(key string, v Value) {
	if i.Fields == nil {
		i.Fields = make(map[string]Value)
	}
	i.Fields[key] = v
}
