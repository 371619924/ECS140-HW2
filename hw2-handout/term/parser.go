package term

import (
	"errors"
	// "strconv"
)

// ErrParser is the error value returned by the Parser if the string is not a
// valid term.
// See also https://golang.org/pkg/errors/#New
// and // https://golang.org/pkg/builtin/#error
var ErrParser = errors.New("parser error")

//
// <start>    ::= <term> | \epsilon
// <term>     ::= ATOM | NUM | VAR | <compound>
// <compound> ::= <functor> LPAR <args> RPAR
// <functor>  ::= ATOM
// <args>     ::= <term> | <term> COMMA <args>
//

// Parser is the interface for the term parser.
// Do not change the definition of this interface.
type Parser interface {
	Parse(string) (*Term, error)
}

type parser struct {
	lex  *lexer
	next *Token
	dag  map[string]*Term
}

// NewParser creates a struct of a type that satisfies the Parser interface.
func NewParser() Parser {
	return &parser{}
}
func (p *parser) Parse(input string) (*Term, error) {
	p.lex = newLexer(input)
	p.dag = make(map[string]*Term)

	err := p.readNext()
	if err != nil {
		return nil, ErrParser
	}

	// <start> ::= \epsilon
	if p.next.typ == tokenEOF {
		return nil, nil
	}

	// <start> ::= <term>
	tm, err := p.parseTerm()
	if err != nil {
		return nil, ErrParser
	}

	// Entire input must be consumed.
	if p.next == nil || p.next.typ != tokenEOF {
		return nil, ErrParser
	}

	return tm, nil
}

func (p *parser) readNext() error {
	tok, err := p.lex.next()
	if err != nil {
		return ErrParser
	}
	p.next = tok
	return nil
}

func (p *parser) parseTerm() (*Term, error) {
	if p.next == nil {
		return nil, ErrParser
	}

	if p.next.typ == tokenNumber {
		lit := p.next.literal
		err := p.readNext()
		if err != nil {
			return nil, err
		}
		return p.share(&Term{Typ: TermNumber, Literal: lit}), nil
	}

	if p.next.typ == tokenVariable {
		lit := p.next.literal
		err := p.readNext()
		if err != nil {
			return nil, err
		}
		return p.share(&Term{Typ: TermVariable, Literal: lit}), nil
	}

	if p.next.typ == tokenAtom {
		lit := p.next.literal
		err := p.readNext()
		if err != nil {
			return nil, err
		}

		atom := p.share(&Term{Typ: TermAtom, Literal: lit})

		// ATOM or <compound>
		if p.next != nil && p.next.typ == tokenLpar {
			return p.parseCompound(atom)
		}
		return atom, nil
	}

	return nil, ErrParser
}

func (p *parser) parseCompound(functor *Term) (*Term, error) {
	// Current token should be '('
	if p.next == nil || p.next.typ != tokenLpar {
		return nil, ErrParser
	}

	err := p.readNext()
	if err != nil {
		return nil, err
	}

	args, err := p.parseArgs()
	if err != nil {
		return nil, err
	}

	if p.next == nil || p.next.typ != tokenRpar {
		return nil, ErrParser
	}

	err = p.readNext()
	if err != nil {
		return nil, err
	}

	tm := &Term{Typ: TermCompound, Literal: "", Functor: functor, Args: args}
	return p.share(tm), nil
}

func (p *parser) parseArgs() ([]*Term, error) {
	first, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	args := []*Term{first}

	if p.next != nil && p.next.typ == tokenComma {
		err = p.readNext()
		if err != nil {
			return nil, err
		}

		rest, err := p.parseArgs()
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(rest); i++ {
			args = append(args, rest[i])
		}
	}

	return args, nil
}

func (p *parser) share(tm *Term) *Term {
	key := termID(tm)
	old, ok := p.dag[key]
	if ok {
		return old
	}
	p.dag[key] = tm
	return tm
}

func termID(tm *Term) string {
	if tm.Typ == TermAtom || tm.Typ == TermNumber || tm.Typ == TermVariable {
		return tm.String()
	}

	// TermCompound
	id := tm.Functor.String()
	id = id + "("
	for i := 0; i < len(tm.Args); i++ {
		if i > 0 {
			id = id + ","
		}
		id = id + termID(tm.Args[i])
	}
	id = id + ")"
	return id
}
