package search

import (
	"strings"
)

func Lex(query string) (tokens []SearchToken) {
	collectingVerbatim := false
	verbatim := strings.Builder{}
	// That's an O(n^3) right?
walker:
	for query != "" {
		for _, kind := range MostTokenKinds {
			if strings.HasPrefix(query, string(kind)) {
				tokens = append(tokens, SearchToken{
					Kind:  kind,
					Value: "",
				})
				query = query[len(kind):]
				goto walker
			}
		}

		if !collectingVerbatim {
			collectingVerbatim = true
			verbatim.Reset()
		}
		verbatim.WriteByte(query[0])
		query = query[1:]
	}
	return nil
}

func Compactify(looseTokens []SearchToken) (tokens []SearchToken) {
	i := 0
	for i < len(looseTokens) {
		if !looseTokens[i].Kind.GotTricks() {
			tokens = append(tokens, looseTokens[i])
			i++
			continue
		}
	}
	return tokens
}
