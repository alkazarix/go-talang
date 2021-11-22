package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	Illegal = "Illegal"
	EOF     = "EOF"

	// Identifiers + Literals
	Identifier = "Identifier"
	Number     = "Number"
	String     = "String"

	// Operators
	Assign   = "="
	Plus     = "+"
	Minus    = "-"
	Bang     = "!"
	Asterisk = "*"
	Slash    = "/"
	Equal    = "=="
	NotEqual = "!="
	Or       = "Or"
	And      = "And"

	LessThanEqual    = "<="
	LessThan         = "<"
	GreaterThanEqual = ">="
	GreaterThan      = ">"

	// Delimiters
	Comma        = ","
	Semicolon    = ";"
	Dot          = "."
	LeftParen    = "("
	RightParen   = ")"
	LeftBrace    = "{"
	RightBrace   = "}"
	LeftBracket  = "["
	RightBracket = "]"

	// Keywords
	Class    = "Class"
	This     = "This"
	Super    = "Super"
	Function = "Function"
	Let      = "Let"
	True     = "True"
	False    = "False"
	Nil      = "Nil"
	If       = "If"
	Else     = "Else"
	While    = "While"
	Return   = "Return"
	Print    = "Print"
)

var keywords = map[string]TokenType{
	"class":  Class,
	"this":   This,
	"super":  Super,
	"fn":     Function,
	"let":    Let,
	"true":   True,
	"false":  False,
	"nil":    Nil,
	"if":     If,
	"else":   Else,
	"while":  While,
	"return": Return,
	"print":  Print,
	"or":     Or,
	"and":    And,
}

func LookupIdentifier(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return Identifier
}
