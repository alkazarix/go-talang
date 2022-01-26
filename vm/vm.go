package vm

import (
	"fmt"

	"github.com/alkazarix/talang/code"
	"github.com/alkazarix/talang/compiler"
	"github.com/alkazarix/talang/valuer"
)

const StackSize = 2048

var True = &valuer.Boolean{Value: true}
var False = &valuer.Boolean{Value: false}

type VM struct {
	instructions code.Instructions
	constants    []valuer.Value
	stack        []valuer.Value
	sp           int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		stack:        make([]valuer.Value, StackSize),
		constants:    bytecode.Constants,
		sp:           0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpOr, code.OpAnd:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterEqual, code.OpGreater:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		}

	}
	return nil
}

func (vm *VM) LastPoppedStackElem() valuer.Value {
	return vm.stack[vm.sp]
}

func (vm *VM) StackTop() valuer.Value {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) push(v valuer.Value) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = v
	vm.sp++

	return nil
}

func (vm *VM) pop() valuer.Value {
	v := vm.stack[vm.sp-1]
	vm.sp--
	return v
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == valuer.NumberType && rightType == valuer.NumberType {
		return vm.executeBinaryNumberOperation(op, left, right)
	}

	if leftType == valuer.BooleanType && rightType == valuer.BooleanType {
		return vm.executeLogicalOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}

func (vm *VM) executeBinaryNumberOperation(
	op code.Opcode,
	left, right valuer.Value,
) error {
	leftValue := left.(*valuer.Number).Value
	rightValue := right.(*valuer.Number).Value

	var result float64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&valuer.Number{Value: result})
}

func (vm *VM) executeLogicalOperation(
	op code.Opcode,
	left, right valuer.Value,
) error {
	leftValue := left.(*valuer.Boolean).Value
	rightValue := right.(*valuer.Boolean).Value

	var result bool

	switch op {
	case code.OpAnd:
		result = leftValue && rightValue
	case code.OpOr:
		result = leftValue || rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&valuer.Boolean{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == valuer.NumberType && right.Type() == valuer.NumberType {
		return vm.executeNumberComparison(op, left, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right.Val() == left.Val()))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right.Val() != left.Val()))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeNumberComparison(
	op code.Opcode,
	left, right valuer.Value,
) error {
	leftValue := left.(*valuer.Number).Value
	rightValue := right.(*valuer.Number).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreater:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	fmt.Printf("operand %s", operand.Inspect())

	switch operand.Val() {
	case true:
		return vm.push(False)
	case false:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != valuer.NumberType {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*valuer.Number).Value
	return vm.push(&valuer.Number{Value: -value})
}

func nativeBoolToBooleanObject(input bool) *valuer.Boolean {
	if input {
		return True
	}
	return False
}
