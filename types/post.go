package types

type Visibility int

const (
	Private Visibility = iota
	Public
)

func VisibilityFromString(s string) Visibility {
	switch s {
	case "public":
		return Public
	default:
		return Private
	}
}

type Post struct {
	ID           int
	CreationTime int64

	URL         string
	Title       string
	Description string
	Visibility  Visibility
}
