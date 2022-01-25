package lexer

import (
	"fmt"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/alkazarix/talang/token"
)

var eof = rune(-1)

type Lexer struct {
	s  *scanner.Scanner
	ch rune
}

func New(input string) *Lexer {
	s := &scanner.Scanner{}
	s.Init(strings.NewReader(input))
	l := &Lexer{
		s: s,
	}
	l.consume()
	return l
}

func (l *Lexer) Lexeme() []token.Token {
	tokens := []token.Token{}
	for !l.isAtEnd() {
		tokens = append(tokens, l.NextToken())
	}
	tokens = append(tokens, l.makeToken(token.EOF, ""))
	return tokens
}

func (l *Lexer) NextToken() (tok token.Token) {

	l.skipWhitespaces()

	switch l.ch {
	case '!':
		if l.match('=') {
			tok = l.makeToken(token.NotEqual, "!=")
		} else {
			tok = l.makeToken(token.Bang, "!")
		}
	case '<':
		if l.match('=') {
			tok = l.makeToken(token.LessThanEqual, "<=")
		} else {
			tok = l.makeToken(token.LessThan, string(l.ch))
		}
	case '>':
		if l.match('=') {
			tok = l.makeToken(token.GreaterThanEqual, ">=")
		} else {
			tok = l.makeToken(token.GreaterThan, string(l.ch))
		}
	case '=':
		if l.match('=') {
			tok = l.makeToken(token.Equal, "==")
		} else {
			tok = l.makeToken(token.Assign, string(l.ch))
		}

	case '*':
		tok = l.makeToken(token.Asterisk, string(l.ch))
	case '/':
		tok = l.makeToken(token.Slash, string(l.ch))
	case ',':
		tok = l.makeToken(token.Comma, string(l.ch))
	case ';':
		tok = l.makeToken(token.Semicolon, string(l.ch))
	case '.':
		tok = l.makeToken(token.Dot, string(l.ch))
	case '(':
		tok = l.makeToken(token.LeftParen, string(l.ch))
	case ')':
		tok = l.makeToken(token.RightParen, string(l.ch))
	case '{':
		tok = l.makeToken(token.LeftBrace, string(l.ch))
	case '}':
		tok = l.makeToken(token.RightBrace, string(l.ch))
	case '[':
		tok = l.makeToken(token.LeftBracket, string(l.ch))
	case ']':
		tok = l.makeToken(token.RightBracket, string(l.ch))
	case '+':
		tok = l.makeToken(token.Plus, string(l.ch))
	case '-':
		tok = l.makeToken(token.Minus, string(l.ch))
	case '"':
		literal, err := l.readString()
		if err != nil {
			tok = l.makeToken(token.Illegal, err.Error())
		} else {
			tok = l.makeToken(token.String, literal)
		}
	case eof:
		tok = l.makeToken(token.EOF, "")
	default:
		if unicode.IsLetter(l.ch) {
			literal := l.readIdentifier()
			tok = l.makeToken(token.LookupIdentifier(literal), literal)
			return
		} else if unicode.IsNumber(l.ch) {
			literal, err := l.readNumber()
			if err != nil {
				tok = l.makeToken(token.Illegal, err.Error())
			} else {
				tok = l.makeToken(token.Number, literal)
			}
			return
		} else {
			tok = l.makeToken(token.Illegal, fmt.Sprintf("unknown token: %s", string(l.ch)))
		}

	}

	l.consume()
	return
}

func (l *Lexer) consume() {
	if l.isAtEnd() {
		return
	}

	ch := l.s.Next()
	if ch == scanner.EOF {
		l.ch = eof
		return
	}
	l.ch = ch
}

func (l *Lexer) peek() rune {
	ch := l.s.Peek()
	return ch
}

func (l *Lexer) isAtEnd() bool {
	return l.ch == eof
}

func (l *Lexer) skipWhitespaces() {
	for unicode.IsSpace(l.ch) {
		l.consume()
	}
}

func (l *Lexer) match(ch rune) bool {
	peek := l.peek()
	if l.isAtEnd() || peek != ch {
		return false
	}
	l.consume()
	return true
}

func (l *Lexer) readString() (string, error) {

	strBuilder := &strings.Builder{}
	l.consume() // start ".
	if l.ch == '"' {
		l.consume()
		return "", nil
	}

	for l.ch != '"' {
		if l.isAtEnd() {
			return "", l.makeError("unterminated string")
		} else {
			strBuilder.WriteRune(l.ch)
		}
		l.consume()
	}
	// end ".
	//	l.consume()
	return strBuilder.String(), nil
}

func (l *Lexer) readNumber() (string, error) {

	strBuilder := &strings.Builder{}
	for unicode.IsNumber(l.ch) {
		strBuilder.WriteRune(l.ch)
		l.consume()
	}

	if l.ch == '.' {
		if !unicode.IsNumber(l.peek()) {
			return strBuilder.String(), nil
		}
		strBuilder.WriteRune(l.ch)
		l.consume()
		for unicode.IsNumber(l.ch) {
			strBuilder.WriteRune(l.ch)
			l.consume()
		}
	}

	return strBuilder.String(), nil
}

func (l *Lexer) readIdentifier() string {
	strBuilder := &strings.Builder{}
	for isAlphaNumeric(l.ch) {
		strBuilder.WriteRune(l.ch)
		l.consume()
	}
	return strBuilder.String()
}

func (l *Lexer) makeToken(ttype token.Type, literal string) token.Token {
	return token.Token{
		Type:     ttype,
		Literal:  literal,
		Position: token.Position{Line: l.s.Line, Column: l.s.Column},
	}
}

func (l *Lexer) makeError(msg string) error {
	return fmt.Errorf("%s %s\n", l.s.Pos().String(), msg)
}

func isAlphaNumeric(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsNumber(ch) || ch == '_'
}
