package vm

import (
	"fmt"
	"testing"

	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/compiler"
	"github.com/alkazarix/talang/lexer"
	"github.com/alkazarix/talang/parser"
	"github.com/alkazarix/talang/valuer"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1;", 1},
		{"2;", 2},
		{"1 + 2;", 3},
	}

	runVmTests(t, tests)
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program, err := parse(tt.input)
		if err != nil {
			t.Fatalf("parsing error: %s", err)
		}

		comp := compiler.New()
		err = comp.Compile(&program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.StackTop()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func parse(input string) (ast.Program, error) {
	l := lexer.New(input)
	tokens := l.Lexeme()
	p := parser.New(tokens)
	return p.Parse()
}

func testExpectedObject(
	t *testing.T,
	expected interface{},
	actual valuer.Value,
) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testNumberValue(float64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	}
}

func testNumberValue(expected float64, actual valuer.Value) error {
	result, ok := actual.(*valuer.Number)
	if !ok {
		return fmt.Errorf("object is not Number. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f",
			result.Value, expected)
	}

	return nil
}
