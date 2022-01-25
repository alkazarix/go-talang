package valuer

import (
	"errors"
	"fmt"
	"time"
)

// Builtin function
type BuiltinFunction func(args ...Value) (Value, error)
type Builtin interface {
	Fn() BuiltinFunction
}

// clock
type Clock struct{}

func (c *Clock) Type() ValueType {
	return BuiltinType
}

func (c *Clock) Inspect() string {
	return "<fn> clock"
}

func (c *Clock) call() {}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) Fn() BuiltinFunction {
	return func(args ...Value) (Value, error) {
		if len(args) != 0 {
			msg := fmt.Sprintf("wrong number of arguments for `clock`. got=%d, want=0", len(args))
			return &Nil{}, errors.New(msg)
		}

		now := time.Now() // current local time
		sec := now.Unix()

		return &Number{Value: float64(sec)}, nil
	}
}

// at
type At struct{}

func (at *At) Type() ValueType {
	return BuiltinType
}

func (at *At) Inspect() string {
	return "<fn> at"
}

func (at *At) call() {}

func (at *At) Arity() int {
	return 2
}

func (at *At) Fn() BuiltinFunction {
	return func(args ...Value) (Value, error) {
		if len(args) != 2 {
			msg := fmt.Sprintf("wrong number of arguments for `at`. got=%d, want=1", len(args))
			return &Nil{}, errors.New(msg)
		}
		if args[0].Type() != ArrayType {
			msg := fmt.Sprintf("argument to `at` must be ARRAY, got %s", args[0].Type())
			return &Nil{}, errors.New(msg)
		}

		if args[1].Type() != NumberType {
			msg := fmt.Sprintf("argument to `at` must be NUMBER, got %s", args[0].Type())
			return &Nil{}, errors.New(msg)
		}

		arr := args[0].(*Array)
		index := int(args[1].(*Number).Value)
		if index < len(arr.Elements) && index >= 0 {
			return arr.Elements[index], nil
		}

		return &Nil{}, nil
	}
}

// len
type Len struct{}

func (l *Len) Type() ValueType {
	return BuiltinType
}

func (l *Len) Inspect() string {
	return "<fn> len"
}

func (l *Len) call() {}

func (l *Len) Arity() int {
	return 1
}

func (l *Len) Fn() BuiltinFunction {
	return func(args ...Value) (Value, error) {
		if len(args) != 1 {
			msg := fmt.Sprintf("wrong number of arguments for `len`. got=%d, want=1", len(args))
			return &Nil{}, errors.New(msg)
		}
		if args[0].Type() == ArrayType {
			l := len(args[0].(*Array).Elements)
			return &Number{Value: float64(l)}, nil
		}

		if args[0].Type() == StringType {
			l := len(args[0].(*String).Value)
			return &Number{Value: float64(l)}, nil
		}

		msg := fmt.Sprintf("argument of `len` must be STRING or ARRAY, got %s", args[0].Type())
		return &Nil{}, errors.New(msg)

	}
}

// push
type Push struct{}

func (p *Push) Type() ValueType {
	return BuiltinType
}

func (p *Push) Inspect() string {
	return "<fn> push"
}

func (p *Push) call() {}

func (p *Push) Arity() int {
	return 2
}

func (p *Push) Fn() BuiltinFunction {
	return func(args ...Value) (Value, error) {
		if len(args) != 2 {
			msg := fmt.Sprintf("wrong number of arguments for `push`. got=%d, want=2", len(args))
			return &Nil{}, errors.New(msg)
		}
		if args[0].Type() != ArrayType {
			msg := fmt.Sprintf("argument to `push` must be ARRAY, got %s", args[0].Type())
			return &Nil{}, errors.New(msg)
		}

		arr := args[0].(*Array)
		length := len(arr.Elements)

		newElements := make([]Value, length+1, length+1)
		copy(newElements, arr.Elements)
		newElements[length] = args[1]

		return &Array{Elements: newElements}, nil
	}
}

// rest
type Rest struct{}

func (r *Rest) Type() ValueType {
	return BuiltinType
}

func (r *Rest) Inspect() string {
	return "<fn> rest"
}

func (r *Rest) call() {}

func (r *Rest) Arity() int {
	return 1
}

func (r *Rest) Fn() BuiltinFunction {
	return func(args ...Value) (Value, error) {
		if len(args) != 1 {
			msg := fmt.Sprintf("wrong number of arguments for `rest`. got=%d, want=1", len(args))
			return &Nil{}, errors.New(msg)
		}
		if args[0].Type() != ArrayType {
			msg := fmt.Sprintf("argument to `rest` must be ARRAY, got %s", args[0].Type())
			return &Nil{}, errors.New(msg)
		}

		arr := args[0].(*Array)
		length := len(arr.Elements)
		if length > 0 {
			newElements := make([]Value, length-1, length-1)
			copy(newElements, arr.Elements[1:length])
			return &Array{Elements: newElements}, nil
		}

		return &Nil{}, nil
	}
}
