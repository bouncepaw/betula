package search

type TokenKind string

const (
	// Both in Token and Node

	Cat      TokenKind = "#"
	Title    TokenKind = "title:"
	Protocol TokenKind = "protocol:"
	URL      TokenKind = "url:"
	Site     TokenKind = "site:"
	Desc     TokenKind = "desc:"
	Or       TokenKind = "|"
	Without  TokenKind = "-"
	Literal  TokenKind = ""

	// In Token only

	Space TokenKind = " "
	Open  TokenKind = "("
	Close TokenKind = ")"

	// In Node only

	And TokenKind = ""

	// The rest

	Quote TokenKind = "\""
)

type Token struct {
	Kind  TokenKind
	Value string
}

type Node struct {
	Kind   TokenKind
	Value  string
	Others []Node
}
