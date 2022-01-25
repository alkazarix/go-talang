package compiler

import (
	"fmt"

	"github.com/alkazarix/talang/token"
)

type CompileError struct {
	message string
}

func NewCompileError(reason string, at *token.Token) *CompileError {
	if at != nil {
		message := fmt.Sprintf("[runtime error] %s (at line: %d, column: %d)", reason, at.Position.Line, at.Position.Column)
		return &CompileError{message: message}
	}
	message := fmt.Sprintf("[runtime error] %s", reason)
	return &CompileError{message: message}
}

func (p *CompileError) Error() string {
	return p.message
}
