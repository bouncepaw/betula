package search

import (
	"testing"
)

func TestLex(t *testing.T) {
	table := map[string][]SearchTokenKind{
		``:                       {},
		` `:                      {},
		`  `:                     {},
		`blac   pink`:            {Verbatim, Space, Verbatim},
		`"shutdown!"`:            {Verbatim},
		`text:"black pink"`:      {Text, Verbatim},
		`text:"text:black pink"`: {Text, Verbatim},
	}
	for input, expectedKinds := range table {
		output := Lex(input)
		for i, token := range output {
			if expectedKinds[i] != token.Kind {
				t.Errorf("Mismatch! Expected: %v. Got: %v.", expectedKinds, output)
			}
		}
	}
}
