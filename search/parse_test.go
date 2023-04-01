package search

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	table := map[string][]Node{
		``:    {},
		`   `: {},
		`this that`: {
			{And, "", []Node{
				{Literal, "this", nil},
				{Literal, "that", nil},
			}},
		},
		`this | that`: {
			{Or, "", []Node{
				{Literal, "this", nil},
				{Literal, "that", nil},
			}},
		},
		`(this that) | "pink venom"`: {
			{Or, "", []Node{
				{And, "", []Node{
					{Literal, "this", nil},
					{Literal, "that", nil},
				}},
				{Literal, "pink venom", nil},
			}},
		},
		`(lovesick girls`: {
			{And, "", []Node{
				{Literal, "lovesick", nil},
				{Literal, "girls", nil},
			}},
		},
		`"hit you with that" - "whistle"`: {
			{Without, "", []Node{
				{Literal, "hit you with that", nil},
				{Literal, "whistle", nil},
			}},
		},
	}

	for input, expected := range table {
		output := Parse(Lex(input))
		for i, node := range output {
			if !reflect.DeepEqual(expected[i], node) {
				t.Errorf("Mismatch! Expected: %v. Got: %v.", expected, output)
			}
		}
	}
}
