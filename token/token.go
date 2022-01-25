package token

type Type string

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Type     Type
	Literal  string
	Position Position
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
	For      = "For"
)

var keywords = map[string]Type{
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
	"for":    For,
}

func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return Identifier
}
