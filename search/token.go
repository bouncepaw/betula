package search

type TokenKind string

const (
	Space    TokenKind = " "
	Or       TokenKind = "|"
	Not      TokenKind = "-"
	Open     TokenKind = "("
	Close    TokenKind = ")"
	Quote    TokenKind = "\""
	Cat      TokenKind = "#"
	Title    TokenKind = "title:"
	Protocol TokenKind = "protocol:"
	URL      TokenKind = "url:"
	Site     TokenKind = "site:"
	Desc     TokenKind = "desc:"
	Literal  TokenKind = ""
)

type Token struct {
	Kind  TokenKind
	Value string
}

var MostTokenKinds = []TokenKind{Or, Not, Open, Close, Cat, Title, Protocol, URL, Site, Desc}

func (kind TokenKind) GotTricks() bool {
	switch kind {
	case Or, Not, Open, Close, Cat, Title, Protocol, URL, Site, Desc:
		return false
	case Space, Quote, Literal:
		return true
	default:
		panic("Magician reveals no secrets")
	}
}
