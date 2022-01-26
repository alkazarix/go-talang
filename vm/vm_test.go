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
		{"1 - 2;", -1},
		{"1 * 2;", 2},
		{"4 / 2;", 2},
		{"50 / 2 * 2 + 10 - 5;", 55},
		{"5 * (2 + 10);", 60},
		{"5 + 5 + 5 + 5 - 10;", 10},
		{"2 * 2 * 2 * 2 * 2;", 32},
		{"5 * 2 + 10;", 20},
		{"5 + 2 * 10;", 25},
		{"5 * (2 + 10);", 60},
		{"-5;", -5},
		{"-10;", -10},
		{"-50 + 100 + -50;", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10;", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{

		{"true;", true},
		{"false;", false},

		{"1 < 2;", true},

		{"1 > 2;", false},
		{"1 < 1;", false},
		{"1 > 1;", false},
		{"1 == 1;", true},
		{"1 != 1;", false},
		{"1 == 2;", false},
		{"1 != 2;", true},

		{"true == true;", true},

		{"false == false;", true},
		{"true == false;", false},
		{"true != false;", true},
		{"false != true;", true},

		{"(1 < 2) == true;", true},

		{"(1 < 2) == false;", false},
		{"(1 > 2) == true;", false},
		{"(1 > 2) == false;", true},

		{"!true;", false},

		{"!false;", true},

		{"!5;", false},
		{"!!true;", true},
		{"!!false;", false},
		{"!!5;", true},
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

		stackElem := vm.LastPoppedStackElem()

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
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
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

func testBooleanObject(expected bool, actual valuer.Value) error {
	result, ok := actual.(*valuer.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
	}

	return nil
}
