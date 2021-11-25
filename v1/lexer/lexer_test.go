package lexer

import (
	"testing"

	"github.com/alkazarix/talang/token"
)

func TestSimpleToken(t *testing.T) {
	input := `
() {}
, . - + ;
/ * !
= == !=
> >=
< <=`
	l := New(input)
	tests := []struct {
		expectTok     token.Type
		expectLiteral string
	}{
		{token.LeftParen, "("},
		{token.RightParen, ")"},
		{token.LeftBrace, "{"},
		{token.RightBrace, "}"},
		{token.Comma, ","},

		{token.Dot, "."},
		{token.Minus, "-"},
		{token.Plus, "+"},
		{token.Semicolon, ";"},
		{token.Slash, "/"},

		{token.Asterisk, "*"},
		{token.Bang, "!"},
		{token.Assign, "="},
		{token.Equal, "=="},
		{token.NotEqual, "!="},

		{token.GreaterThan, ">"},
		{token.GreaterThanEqual, ">="},
		{token.LessThan, "<"},
		{token.LessThanEqual, "<="},
	}

	for i, test := range tests {
		tok := l.NextToken()
		if test.expectTok != tok.Type {
			t.Fatalf("test [%d]: expected token is %s. got %s", i, test.expectTok, tok.Type)
		}
	}

	tok := l.NextToken()
	if token.EOF != tok.Type {
		t.Fatalf("expected token is EOF. got %s", tok.Type)
	}
}

func TestIdentifier(t *testing.T) {
	input := `
abc 		xyz 		a123 		A_123			X_x_
and 		class 	else 		false 		fn
if 			nil 		or 			print     return 	
super 	this 		true		let       while
`
	tests := []struct {
		expectTok     token.Type
		expectLiteral string
	}{
		{token.Identifier, "abc"},
		{token.Identifier, "xyz"},
		{token.Identifier, "a123"},
		{token.Identifier, "A_123"},
		{token.Identifier, "X_x_"},

		{token.And, "and"},
		{token.Class, "class"},
		{token.Else, "else"},
		{token.False, "false"},
		{token.Function, "fn"},

		{token.If, "if"},
		{token.Nil, "nil"},
		{token.Or, "or"},
		{token.Print, "print"},

		{token.Return, "return"},
		{token.Super, "super"},
		{token.This, "this"},
		{token.True, "true"},
		{token.Let, "let"},

		{token.While, "while"},
	}
	l := New(input)

	for i, test := range tests {
		token := l.NextToken()

		if token.Type != test.expectTok {
			t.Fatalf("test [%d]: expected token is %s. got %s", i, test.expectTok, token.Type)
		}

		if token.Literal != test.expectLiteral {
			t.Fatalf("test [%d]: expected literal is %s. got %s", i, test.expectLiteral, token.Literal)
		}
	}
}

func TestNumber(t *testing.T) {
	input := `
0
01
0123
123
123.0
0.123
123.456`
	l := New(input)
	tests := []string{
		"0", "01", "0123", "123", "123.0", "0.123", "123.456",
	}

	for i, test := range tests {
		tok := l.NextToken()
		if tok.Type != token.Number {
			t.Fatalf("test [%d]: expected token is number. got %s", i, tok.Type)
		}

		if tok.Literal != test {
			t.Fatalf("test [%d]: expected literal is %q. got %q", i, test, tok.Literal)
		}
	}
}

func TestString(t *testing.T) {
	tests := []string{
		"",
		"abc xyz",
		"字符串",
	}
	input := `"" "abc xyz" "字符串"`

	l := New(input)
	for _, expected := range tests {
		tok := l.NextToken()

		if tok.Type != token.String {
			t.Fatalf("expected token is string. got %s", tok.Type)
		}

		if tok.Literal != expected {
			t.Fatalf("expected literal is %q. got %q", expected, tok.Literal)
		}
	}
}
