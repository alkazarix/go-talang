package interpreter

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/alkazarix/talang/lexer"
	"github.com/alkazarix/talang/parser"
	"github.com/alkazarix/talang/valuer"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5;", 5},

		{"10;", 10},
		{"-5;", -5},
		{"-10;", -10},

		{"5 + 5 + 5 + 5 - 10;", 10},

		{"2 * 2 * 2 * 2 * 2;", 32},

		{"-50 + 100 + -50;", 0},

		{"5 * 2 + 10;", 20},

		{"5 + 2 * 10;", 25},

		{"20 + 2 * -10;", 0},

		{"50 / 2 * 2 + 10;", 60},
		{"3 * 3 * 3 + 10;", 37},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err.Error())
		}
		testNumberValue(t, evaluated, float64(tt.expected))
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true ;", false},
		{"!false ;", true},
		{"!5 ;", false},
		{"!!true ;", true},
		{"!!false ;", false},
		{"!!5 ;", true},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testBooleanValue(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},

		{"1 < 2 ;", true},
		{"1 > 2 ;", false},
		{"1 < 1 ;", false},
		{"1 > 1 ;", false},

		{"1 == 1 ;", true},
		{"1 != 1 ;", false},
		{"1 == 2 ;", false},
		{"1 != 2 ;", true},

		{"true == true ;", true},

		{"false == false ;", true},
		{"true == false ;", false},
		{"true != false ;", true},
		{"false != true ;", true},

		{"(1 < 2) == true ;", true},

		{"(1 < 2) == false ;", false},
		{"(1 > 2) == true ;", false},
		{"(1 > 2) == false ;", true},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testBooleanValue(t, evaluated, tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!" ;`

	evaluated, err := testEval(input)
	if err != nil {
		t.Fatalf("parser error: %s", err)
	}
	str, ok := evaluated.(*valuer.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!" ;`

	evaluated, err := testEval(input)
	if err != nil {
		t.Fatalf("parser error: %s", err)
	}
	str, ok := evaluated.(*valuer.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestArray(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3];"

	evaluated, err := testEval(input)
	if err != nil {
		t.Fatalf("parser error: %s", err)
	}
	result, ok := evaluated.(*valuer.Array)
	if !ok {
		t.Fatalf("object is not %T. got=%T (%+v)", valuer.Array{}, evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testNumberValue(t, result.Elements[0], 1)
	testNumberValue(t, result.Elements[1], 4)
	testNumberValue(t, result.Elements[2], 6)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`let l = len("four"); l;`, 4},
		{`let l = len([1, 2, 3]); l;`, 3},
		/*
			{`let l = len("four"); l;`, 4},
			{`let l = len("hello world"); l;`, 11},
			{`let l = len([1, 2, 3]); l;`, 3},
			{`let l = len([]); l;`, 0},
		*/
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}

		switch expected := tt.expected.(type) {
		case int:
			testNumberValue(t, evaluated, float64(expected))
		case nil:
			testNilValue(t, evaluated)
		case []int:
			array, ok := evaluated.(*valuer.Array)
			if !ok {
				t.Errorf("obj not Array. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if len(array.Elements) != len(expected) {
				t.Errorf("wrong num of elements. want=%d, got=%d",
					len(expected), len(array.Elements))
				continue
			}

			for i, expectedElem := range expected {
				testNumberValue(t, array.Elements[i], float64(expectedElem))
			}
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 3; a = 2; a;", 2},
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testNumberValue(t, evaluated, float64(tt.expected))
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10; }", 10},
		{"if (false) { 10; }", nil},
		{"if (1) { 10; }", 10},
		{"if (1 < 2) { 10; }", 10},
		{"if (1 > 2) { 10; }", nil},
		{"if (1 > 2) { 10; } else { 20; }", 20},
		{"if (1 < 2) { 10; } else { 10; }", 10},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		integer, ok := tt.expected.(int)
		if ok {
			testNumberValue(t, evaluated, float64(integer))
		} else {
			testNilValue(t, evaluated)
		}
	}
}

func TestEvalWhileStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			`let a = 0;
	while (a < 3) {
		a = a + 1;
	}
	a;`, 3},
		{
			`let a = 0;
		while (false) {
			a = a + 1;
		}
		a;`, 0},
	}
	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testNumberValue(t, evaluated, float64(tt.expected))
	}
}

func TestForStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			`let b = 0;
		for (let a = 0; a < 3; a = a + 1)  {
			b = a;
	}
	b;`, 2},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testNumberValue(t, evaluated, float64(tt.expected))
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"fn identity(x) { 5; } identity(5);", nil},
		{"fn identity(x) { return x; } identity(5);", 5},
		{"fn double(x) { return x * 2; } double(5);", 10},
		{"fn add(x, y) { let z = x +y; return z; } add(5, 5);", 10},
		{"fn add(x, y) { let z = x +y; return z; } add(5 + 5, add(5, 5));", 20},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		integer, ok := tt.expected.(int)
		if ok {
			testNumberValue(t, evaluated, float64(integer))
		} else {
			testNilValue(t, evaluated)
		}
	}
}

func TestFunctionClosure(t *testing.T) {
	input := `
	fn gen(x) {
		let a = 0;
		fn inner(y) {
			a = a + 1;
			return a + x + y;
		}
		return inner;
	}
	let g = gen(1);
	g(1);`

	expected := 3
	evaluated, err := testEval(input)
	if err != nil {
		t.Fatalf("parser error: %s", err)
	}

	testNumberValue(t, evaluated, float64(expected))
}

func TestEvalClass(t *testing.T) {

	klass := `class A {
		m() {
			return "a.m";
		}
	}

	class B {
		m() {
			return "b.m";
		}
	}`

	tests := []struct {
		input    string
		expected interface{}
	}{
		{klass + " let a = A(); a.y = 1; a.y;", 1},

		{klass + " let a = A(); let f = a.m(); f;", "a.m"},
		{klass + " let a = A(); a.y = 1; let b = B(); b.m();", "b.m"},
		{klass + " let a = A(); a.y = 1;  let b = B(); b.x = a; b.x.y;", 1},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		switch expected := tt.expected.(type) {
		case int:
			testNumberValue(t, evaluated, float64(expected))
		case string:
			testStringValue(t, evaluated, expected)
		default:
			t.Fatalf("expected type for test")
		}
	}

}

func TestEvalInstance(t *testing.T) {

	klass := `class A {
		init(y) {
			this.y = y;
		}
		m() {
			return this.x;
		}
	}
	`

	tests := []struct {
		input    string
		expected int
	}{
		{klass + " let a = A(2); a.y;", 2},
		{klass + " let a = A(2); a.x = 1; a.m();", 1},
	}

	for _, tt := range tests {
		evaluated, err := testEval(tt.input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		testNumberValue(t, evaluated, float64(tt.expected))
	}
}

func TestPrintStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"print 5;", []string{"5"}},
		{"print 1; print 2; print 3;", []string{"1", "2", "3"}},
		{`print "hello";`, []string{"hello"}},
	}

	for _, tt := range tests {
		testPrint(t, tt.input, tt.expected)
	}
}

func testEval(input string) (valuer.Value, error) {
	l := lexer.New(input)
	tokens := l.Lexeme()
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, err
	}

	interpreter := New()
	return interpreter.Evaluate(&program)
}

func testNumberValue(t *testing.T, value valuer.Value, expected float64) bool {
	result, ok := value.(*valuer.Number)
	if !ok {
		t.Errorf("object is not Number. got=%T (%+v)", value, value)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}

	return true
}

func testBooleanValue(t *testing.T, value valuer.Value, expected bool) bool {
	result, ok := value.(*valuer.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", value, value)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}

	return true
}

func testStringValue(t *testing.T, value valuer.Value, expected string) bool {
	result, ok := value.(*valuer.String)
	if !ok {
		t.Errorf("object is not a String. got=%T (%+v)", value, value)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}

	return true
}

func testNilValue(t *testing.T, value valuer.Value) bool {
	if value != Nil {
		t.Errorf("value is not NULL. got=%T (%+v)", value, value)
		return false
	}
	return true
}

func testPrint(t *testing.T, input string, expected []string) {
	// reset environment before interpereting.
	s := captureStdout(func() {
		_, err := testEval(input)
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
	})
	out := splitByLine(s)
	if len(out) != len(expected) {
		t.Errorf("should get %d outputs. got %d", len(expected), len(out))
		return
	}
	for i, s := range out {
		if s != expected[i] {
			t.Errorf("expected output is %s. got %s", expected[i], s)
			return
		}
	}
}

// https://stackoverflow.com/a/47281683
func captureStdout(fn func()) string {
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w

	fn()

	ch := make(chan string)
	go func() {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		ch <- string(b)
	}()
	w.Close()
	os.Stdout = rescueStdout
	s := <-ch
	return s
}

func splitByLine(s string) []string {
	s = strings.TrimSpace(s)
	return strings.Split(s, "\n")
}
