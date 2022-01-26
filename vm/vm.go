package vm

import (
	"fmt"

	"github.com/alkazarix/talang/code"
	"github.com/alkazarix/talang/compiler"
	"github.com/alkazarix/talang/valuer"
)

const StackSize = 2048
const GlobalsSize = 65536

var True = &valuer.Boolean{Value: true}
var False = &valuer.Boolean{Value: false}
var Nil = &valuer.Nil{}

type VM struct {
	instructions code.Instructions
	constants    []valuer.Value
	stack        []valuer.Value
	sp           int // Always points to the next value. Top of stack is stack[sp-1]
	globals      []valuer.Value
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		stack:        make([]valuer.Value, StackSize),
		constants:    bytecode.Constants,
		sp:           0,
		globals:      make([]valuer.Value, GlobalsSize),
	}
}

func (vm *VM) Run() error {
	fmt.Printf(vm.instructions.String())
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

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpNil:
			err := vm.push(Nil)
			if err != nil {
				return err
			}

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

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				fmt.Printf("not truthy")
				ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
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

	if leftType == valuer.StringType && rightType == valuer.StringType {
		return vm.executeBinaryStringOperation(op, left, right)
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

func (vm *VM) executeBinaryStringOperation(
	op code.Opcode,
	left, right valuer.Value,
) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*valuer.String).Value
	rightValue := right.(*valuer.String).Value

	return vm.push(&valuer.String{Value: leftValue + rightValue})
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

func (vm *VM) buildArray(startIndex, endIndex int) valuer.Value {
	elements := make([]valuer.Value, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &valuer.Array{Elements: elements}
}

func nativeBoolToBooleanObject(input bool) *valuer.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(val valuer.Value) bool {
	switch val := val.(type) {

	case *valuer.Boolean:
		return val.Value

	case *valuer.Nil:
		return false

	default:
		return true
	}
}
