package valuer

import "fmt"

type ValueType string

const (
	NumberType  = "Number"
	BooleanType = "Boolean"
	NilType     = "Nil"
	ErrorType   = "Error"
	StringType  = "String"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

// number value
type Number struct {
	Value float64
}

func (n *Number) Type() ValueType { return NumberType }
func (n *Number) Inspect() string { return fmt.Sprintf("%f", n.Value) }

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

// error value
type Error struct {
	Error error
}

func (e *Error) Type() ValueType { return ErrorType }
func (e *Error) Inspect() string { return e.Error.Error() }
