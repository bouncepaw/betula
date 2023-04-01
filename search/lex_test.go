package search

import (
	"testing"
)

func TestLex(t *testing.T) {
	table := map[string][]Token{
		``:                       {},
		` `:                      {},
		`  `:                     {},
		`black   pink `:          {{Literal, "black"}, {Space, ""}, {Literal, "pink"}},
		`"shutdown!"`:            {{Literal, "shutdown!"}},
		`desc:"black pink"`:      {{Desc, ""}, {Literal, "black pink"}},
		`desc:"text:black pink"`: {{Desc, ""}, {Literal, "text:black pink"}},
		`black"pink`:             {{Literal, "black\"pink"}},
		`black(pink)`:            {{Literal, "black"}, {Open, ""}, {Literal, "pink"}, {Close, ""}},
		`site:url:"pink venom"`:  {{Site, ""}, {URL, ""}, {Literal, "pink venom"}},
		`this|that`:              {{Literal, "this"}, {Or, ""}, {Literal, "that"}},
		`"when we" "pull up"`:    {{Literal, "when we"}, {Space, ""}, {Literal, "pull up"}},
		`#lovesick #girls`:       {{Cat, ""}, {Literal, "lovesick"}, {Space, ""}, {Cat, ""}, {Literal, "girls"}},
	}
	for input, expected := range table {
		output := Lex(input)
		for i, token := range output {
			if expected[i] != token {
				t.Errorf("Mismatch! Expected: %v. Got: %v.", expected, output)
			}
		}
	}
}
