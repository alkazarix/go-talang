package parser

import (
	"testing"

	"github.com/alkazarix/talang/ast"
	"github.com/alkazarix/talang/lexer"
)

type parserTest struct {
	input    string
	expected string
}

func TestParseExpression(t *testing.T) {
	tests := []parserTest{

		{
			input:    "123 - -456 - 789",
			expected: "((123 - (-456)) - 789)",
		},
		{
			input:    "1 >= 2 == !false",
			expected: "((1 >= 2) == (!false))",
		},

		{
			input:    "123 - 456 * 789 / 123",
			expected: "(123 - ((456 * 789) / 123))",
		},
	}

	checkExpr(t, tests)
}

func TestParseLogicExpr(t *testing.T) {
	tests := []parserTest{
		{
			input:    "true or false",
			expected: "(true Or false)",
		},
		{
			input:    "true and false",
			expected: "(true And false)",
		},

		{
			input:    "false or (2 == 2)",
			expected: "(false Or ((2 == 2)))",
		},

		{
			input:    "a == 2 and false",
			expected: "((a == 2) And false)",
		},
	}
	checkExpr(t, tests)
}

func TestParseGetExpr(t *testing.T) {
	tests := []parserTest{
		{
			input:    "x.y",
			expected: "x.y",
		},
		{
			input:    "x.y.z",
			expected: "x.y.z",
		},
	}
	checkExpr(t, tests)
}

func TestParseArrayExpr(t *testing.T) {
	tests := []parserTest{

		{
			input:    "[1, 2, 3]",
			expected: "[1, 2, 3]",
		},

		{
			input:    "[]",
			expected: "[]",
		},
	}
	checkExpr(t, tests)
}

func TestParseSetExpr(t *testing.T) {
	tests := []parserTest{
		{
			input:    "x.y = 1",
			expected: "x.y = 1",
		},
		{
			input:    "x.y.z = 1",
			expected: "x.y.z = 1",
		},
	}
	checkExpr(t, tests)
}

func TestParsePrintStatement(t *testing.T) {
	input := `let a = 0;
		a = a + 10;
		let b = a = a + 100;
		print a;`
	expected := []string{
		"let a = 0;",
		"a = (a + 10);",
		"let b = a = (a + 100);",
		"print a;",
	}
	checkAst(t, input, expected)
}

func TestParseConditionStatement(t *testing.T) {
	tests := []parserTest{
		{
			input:    `if (a < 1 ) { print a;}`,
			expected: `if ((a < 1)) { print a; }`,
		},
		{
			input:    `if (a < 1 ) { print a;}`,
			expected: `if ((a < 1)) { print a; }`,
		},

		{
			input: `if (a) {
					print a;
				} else {
					print 0;
				}`,
			expected: `if (a) { print a; } else { print 0; }`,
		},
		{
			input: `while (a < 2) {
					print a;
				}`,
			expected: "while ((a < 2)) { print a; }",
		},
		{
			input: `for (let a = 1; a < 2; a = a + 1) {
				print a;
			}`,
			expected: block("let a = 1;while ((a < 2)) { { print a; }a = (a + 1); }"),
		},
		{
			input: `for (; a < 2; a = a + 1) {
				print a;
			}`,
			expected: "while ((a < 2)) { { print a; }a = (a + 1); }",
		},

		{
			input: `for (;;a = a + 1) {
					print a;
				}`,
			expected: "while (true) { { print a; }a = (a + 1); }",
		},

		{
			input: `for (;;) {
					print a;
				}`,
			expected: "while (true) { print a; }",
		},
	}
	for i, test := range tests {
		p := newParser(t, test.input)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("test [%d]: parse failed. error: %s", i, err.Error())
		}

		statements := program.Statements
		if len(statements) != 1 {
			t.Fatalf("test [%d]: should have 1 statement. got %d", i, len(statements))
		}
		s := statements[0].String()
		if s != test.expected {
			t.Fatalf("test [%d]: expected content is %q. got %q", i, test.expected, s)
		}
	}
}

func TestParseFunction(t *testing.T) {

	input := `fn t() {
		        print a;
		        return a;
				}
				fn t1(x,y,z) {
		        print a;
		        return a;
		    }
		    fn t2(x,y,z) {
		        print a;
		        return;
			}
			t2(1,2,3);`

	body := "print a;return a;"
	expected := []string{
		"fn t() " + block(body),
		"fn t1(x, y, z) " + block(body),
		"fn t2(x, y, z) " + block("print a;return ;"),
		"t2(1, 2, 3);",
	}
	checkAst(t, input, expected)
}

func TestParseClass(t *testing.T) {
	input := `
	class A {}
	class B {
		test () {
			return this.x;
		}
	}
	class C {
		test () {
			return "test";
		}
	}
	class D < C  {}
	`

	expected := []string{
		"class A",
		"class B",
		"class C",
		"class D",
	}
	checkAst(t, input, expected)
}

func parseExpression(p *Parser) (expr ast.Expr, err error) {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(error); ok {
				err = parseErr
				expr = nil
			} else {
				panic(r)
			}
		}
	}()
	expr = p.expression()
	return expr, nil
}

func newParser(t *testing.T, input string) *Parser {
	l := lexer.New(input)
	tokens := l.Lexeme()

	for _, token := range tokens {
		t.Log(token.Literal, token.Type)
	}

	return New(tokens)
}

func checkExpr(t *testing.T, tests []parserTest) {

	for i, test := range tests {
		p := newParser(t, test.input)
		expr, err := parseExpression(p)

		if err != nil {
			t.Fatalf("test [%d] parse failed. error is: %s", i, err.Error())
		}

		if expr.String() != test.expected {

			t.Fatalf("test [%d] expected expression is %q. got %q.", i, test.expected, expr.String())
		}

	}
}

func checkAst(t *testing.T, input string, expected []string) {
	p := newParser(t, input)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse failed. error: %s", err.Error())
	}
	statements := program.Statements
	if len(statements) != len(expected) {
		t.Fatalf("length of statements should be %d. got %d", len(expected), len(statements))
	}
	for i, stmt := range statements {
		if stmt.String() != expected[i] {
			t.Errorf("test [%d]: expected text is %q. got %q", i, expected[i], stmt.String())
		}
	}
}

func block(s string) string {
	return "{ " + s + " }"
}
