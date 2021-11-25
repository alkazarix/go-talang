package interpreter

import (
	"fmt"

	"github.com/alkazarix/talang/token"
)

type RuntimeError struct {
	message string
}

func NewRuntimeError(reason string, at token.Token) *RuntimeError {
	message := fmt.Sprintf("[runtime error] %s (at line: %d, column: %d)", reason, at.Position.Line, at.Position.Column)
	return &RuntimeError{message: message}
}

func (p *RuntimeError) Error() string {
	return p.message
}
