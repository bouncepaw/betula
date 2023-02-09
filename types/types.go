// Package types provides common data types used across all Betula, all conveniently collected in a single box for resolving import cycles nicely.
package types

import (
	"git.sr.ht/~bouncepaw/mycomarkup/v5/util"
	"html/template"
	"net/url"
	"strings"
)

// Visibility determines where the post is seen.
type Visibility int

const (
	// Private posts are only seen by the author.
	Private Visibility = iota
	// Public posts are seen by everyone, and are federated.
	Public
)

// VisibilityFromString turns a string into a Visbility.
func VisibilityFromString(s string) Visibility {
	switch s {
	case "private":
		return Private
	default:
		return Public
	}
}

// Post is a link, along with some data.
type Post struct {
	// ID is a unique identifier of the post. Do not set this field by yourself.
	ID int
	// CreationTime is UNIX seconds. Do not set this field by yourself.
	CreationTime int64
	// Categories are categories of this post. Do not set this field by yourself.
	Categories []Category

	// URL is a URL with any protocol.
	URL string
	// Title is a name for the link.
	Title string
	// Description is a Mycomarkup-formatted document. Currently, just unescaped plain text.
	Description string
	// Visibility sets who can see the post.
	Visibility Visibility
}

// FixMixUp returns a copy of post. Its Title and URL are swapped, iff Title is a valid URL, and URL is not a valid URL.
func (post Post) FixMixUp() Post {
	_, err := url.ParseRequestURI(post.URL)
	// If URL is a valid URL:
	if err == nil {
		return post
	}

	_, err = url.ParseRequestURI(post.Title)
	// If Title is not a valid URL
	if err != nil {
		return post
	}

	// At this point, URL is not a valid URL, Title is valid URL. Swap:
	post.Title, post.URL = post.URL, post.Title
	return post
}

type Category struct {
	Name      string
	PostCount uint
}

func CanonicalCategoryName(rawName string) string {
	a := util.CanonicalName(rawName)
	b := strings.ReplaceAll(a, ",", "")
	c := strings.ReplaceAll(b, "/", "")
	return c
}

func JoinCategories(cats []Category) string {
	var buf strings.Builder
	for i, cat := range cats {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(cat.Name)
	}
	return buf.String()
}

func SplitCategories(commaSeparated string) []Category {
	catNames := strings.Split(commaSeparated, ",")
	cats := make([]Category, len(catNames))
	for i, catName := range catNames {
		cats[i] = Category{
			Name: CanonicalCategoryName(catName),
		}
	}
	return cats
}

type Settings struct {
	NetworkPort               uint
	SiteTitle                 template.HTML
	SiteDescriptionMycomarkup string
}
