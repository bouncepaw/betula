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

func (s *lexState) nextToken(buf *strings.Builder) (token *SearchToken, done bool) {
	switch {
	case s.collectingQuoted:
		switch {
		case len(s.input) == 0: // end of input
			return &SearchToken{Kind: Verbatim, Value: buf.String()}, true

		case s.input[0] == '"': // end of quoted string
			s.dropBytes(1)
			defer buf.Reset()
			s.collectingQuoted = false
			return &SearchToken{Kind: Verbatim, Value: buf.String()}, false

		default:
			buf.WriteByte(s.input[0])
			s.dropBytes(1)
			return nil, false
		}

	case s.collectingWord:
		switch {
		case len(s.input) == 0: // end of input
			return &SearchToken{Kind: Verbatim, Value: buf.String()}, true

		case strings.ContainsRune(" ()|", rune(s.input[0])):
			// no dropBytes
			s.collectingWord = false
			defer buf.Reset()
			return &SearchToken{Kind: Verbatim, Value: buf.String()}, false

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
			return &SearchToken{Kind: Space}, false
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
			for _, tok := range MostTokenKinds {
				if strings.HasPrefix(s.input, string(tok)) {
					defer s.dropBytes(len(tok))
					return &SearchToken{
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

func Lex(query string) (tokens []SearchToken) {
	var (
		tok  *SearchToken
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
