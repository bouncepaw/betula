package search

import (
	"testing"
)

func TestLex(t *testing.T) {
	table := map[string][]SearchTokenKind{
		``:                       {},
		` `:                      {Space},
		`  `:                     {Space, Space},
		`"shutdown!"`:            {Quote, Verbatim, Quote},
		`text:"black pink"`:      {Text, Quote, Verbatim, Space, Verbatim, Quote},
		`text:"text:black pink"`: {Text, Quote, Text, Verbatim, Space, Verbatim, Quote},
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

func TestCompactify(t *testing.T) {
	table := map[string][]SearchTokenKind{
		``:                       {},
		` `:                      {Space},
		`  `:                     {Space},
		`"shutdown!"`:            {Verbatim},
		`text:"black pink"`:      {Text, Verbatim},
		`text:"text:black pink"`: {Text, Verbatim},
	}
	for input, expectedKinds := range table {
		output := Compactify(Lex(input))
		for i, token := range output {
			if expectedKinds[i] != token.Kind {
				t.Errorf("Mismatch! Expected: %v. Got: %v.", expectedKinds, output)
			}
		}
	}
}
