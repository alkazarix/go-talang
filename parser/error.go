package parser

type ParseError struct {
	s string
}

func (p *ParseError) Error() string {
	return p.s
}
