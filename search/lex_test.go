package search

import (
	"testing"
)

func TestLex(t *testing.T) {
	table := map[string][]SearchToken{
		``:                       {},
		` `:                      {},
		`  `:                     {},
		`black   pink `:          {{Verbatim, "black"}, {Space, ""}, {Verbatim, "pink"}},
		`"shutdown!"`:            {{Verbatim, "shutdown!"}},
		`text:"black pink"`:      {{Text, ""}, {Verbatim, "black pink"}},
		`text:"text:black pink"`: {{Text, ""}, {Verbatim, "text:black pink"}},
		`black"pink`:             {{Verbatim, "black\"pink"}},
		`black(pink)`:            {{Verbatim, "black"}, {Open, ""}, {Verbatim, "pink"}, {Close, ""}},
		`site:url:"pink venom"`:  {{Site, ""}, {URL, ""}, {Verbatim, "pink venom"}},
		`this|that`:              {{Verbatim, "this"}, {Or, ""}, {Verbatim, "that"}},
		`"when we" "pull up"`:    {{Verbatim, "when we"}, {Space, ""}, {Verbatim, "pull up"}},
	}
	for input, expectedKinds := range table {
		output := Lex(input)
		for i, token := range output {
			if expectedKinds[i] != token {
				t.Errorf("Mismatch! Expected: %v. Got: %v.", expectedKinds, output)
			}
		}
	}
}
