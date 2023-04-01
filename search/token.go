package search

type SearchTokenKind string

const (
	Space    SearchTokenKind = " "
	Or       SearchTokenKind = "|"
	Not      SearchTokenKind = "-"
	Open     SearchTokenKind = "("
	Close    SearchTokenKind = ")"
	Quote    SearchTokenKind = "\""
	Cat      SearchTokenKind = "cat:"
	Title    SearchTokenKind = "title:"
	Protocol SearchTokenKind = "protocol:"
	URL      SearchTokenKind = "url:"
	Site     SearchTokenKind = "site:"
	Text     SearchTokenKind = "text:"
	Verbatim SearchTokenKind = ""
)

type SearchToken struct {
	Kind  SearchTokenKind
	Value string
}

var MostTokenKinds = []SearchTokenKind{Space, Or, Not, Open, Close, Quote, Cat, Title, Protocol, URL, Site, Text}

func (kind SearchTokenKind) GotTricks() bool {
	switch kind {
	case Or, Not, Open, Close, Cat, Title, Protocol, URL, Site, Text:
		return false
	case Space, Quote, Verbatim:
		return true
	default:
		panic("Magician reveals no secrets")
	}
}
