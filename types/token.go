package types

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
