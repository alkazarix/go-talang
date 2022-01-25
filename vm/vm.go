package vm

import (
	"fmt"

	"github.com/alkazarix/talang/code"
	"github.com/alkazarix/talang/compiler"
	"github.com/alkazarix/talang/valuer"
)

const StackSize = 2048

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
		}
	}
	return nil
}

func (vm *VM) push(v valuer.Value) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = v
	vm.sp++

	return nil
}
