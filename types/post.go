package types

type Post struct {
	ID           int
	URL          string
	Title        string
	Description  string
	IsPublic     bool
	CreationTime int64
}
