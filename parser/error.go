package parser

type parseError struct {
	s string
}

func (p *parseError) Error() string {
	return p.s
}
