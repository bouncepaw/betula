package search

import (
	"strings"
)

type lexState struct {
	collectingQuoted bool
	collectingWord   bool
	collectingSpaces bool
	input            string
}

func (s *lexState) dropBytes(n int) {
	s.input = s.input[n:]
}

func (s *lexState) nextToken(buf *strings.Builder) (token *Token, done bool) {
	switch {
	case s.collectingQuoted:
		switch {
		case len(s.input) == 0: // end of input
			return &Token{Kind: Literal, Value: buf.String()}, true

		case s.input[0] == '"': // end of quoted string
			s.dropBytes(1)
			defer buf.Reset()
			s.collectingQuoted = false
			return &Token{Kind: Literal, Value: buf.String()}, false

		default:
			buf.WriteByte(s.input[0])
			s.dropBytes(1)
			return nil, false
		}

	case s.collectingWord:
		switch {
		case len(s.input) == 0: // end of input
			return &Token{Kind: Literal, Value: buf.String()}, true

		case strings.ContainsRune(" ()|", rune(s.input[0])):
			// no dropBytes
			s.collectingWord = false
			defer buf.Reset()
			return &Token{Kind: Literal, Value: buf.String()}, false

		default:
			buf.WriteByte(s.input[0])
			s.dropBytes(1)
			return nil, false
		}

	case s.collectingSpaces:
		switch {
		case len(s.input) == 0:
			return nil, true

		case s.input[0] == ' ':
			s.dropBytes(1)
			return nil, false

		default:
			s.collectingSpaces = false
			return &Token{Kind: Space}, false
		}

	default:
		switch {
		case len(s.input) == 0: // end of input
			return nil, true

		case s.input[0] == ' ':
			s.dropBytes(1)
			s.collectingSpaces = true
			return nil, false

		case s.input[0] == '"':
			s.dropBytes(1)
			s.collectingQuoted = true
			return nil, false

		default:
			for _, tok := range []TokenKind{Or, Not, Open, Close, Cat, Title, Protocol, URL, Site, Desc} {
				if strings.HasPrefix(s.input, string(tok)) {
					defer s.dropBytes(len(tok))
					return &Token{
						Kind:  tok,
						Value: "",
					}, false
				}
			}
			s.collectingWord = true
			return nil, false
		}
	}
}

func Lex(query string) (tokens []Token) {
	var (
		tok  *Token
		done bool
		s    = lexState{input: query}
		buf  strings.Builder
	)
	for !done {
		tok, done = s.nextToken(&buf)
		if tok != nil {
			tokens = append(tokens, *tok)
		}
	}
	return tokens
}
