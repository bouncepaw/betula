package search

func Parse(tokens []Token) []Node {
	return subparse(tokens, 0)
}

func subparse(tokens []Token, from int) []Node {
	return nil
}

type MarkovRule struct {
	Pattern  []TokenKind
	Callback func(tokens []Token, offset int)
}
