// Package types provides common data types used across all Betula, all conveniently collected in a single box for resolving import cycles nicely.
package types

import (
	"git.sr.ht/~bouncepaw/mycomarkup/v5/util"
	"html/template"
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
	// CreationTime is like 2006-01-02 15:04:05.
	CreationTime string
	// Tags are the tags this post has. Do not set this field by yourself.
	Tags []Tag

	// URL is a URL with any protocol.
	URL string
	// Title is a name for the link.
	Title string
	// Description is a Mycomarkup-formatted document. Currently, just unescaped plain text.
	Description string
	// Visibility sets who can see the post.
	Visibility Visibility
}

type Tag struct {
	Name        string
	Description string
	PostCount   uint
}

func CanonicalTagName(rawName string) string {
	a := util.CanonicalName(rawName)
	b := strings.ReplaceAll(a, ",", "")
	c := strings.ReplaceAll(b, "/", "")
	return c
}

func JoinTags(tags []Tag) string {
	var buf strings.Builder
	for i, tag := range tags {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(tag.Name)
	}
	return buf.String()
}

func SplitTags(commaSeparated string) []Tag {
	tagNames := strings.Split(commaSeparated, ",")
	tags := make([]Tag, len(tagNames))
	for i, tagName := range tagNames {
		tags[i] = Tag{
			Name: CanonicalTagName(tagName),
		}
	}
	return tags
}

type Settings struct {
	NetworkPort uint
	// SiteName is a plaintext name of the site.
	SiteName string
	// SiteTitle is a hypertext title shown in the top left corner, in a <h1>.
	SiteTitle                 template.HTML
	SiteDescriptionMycomarkup string
	SiteURL                   string
}

// not really a type:

const TimeLayout = "2006-01-02 15:04:05"

func StripCommonProtocol(a string) string {
	b := strings.TrimPrefix(a, "https://")
	c := strings.TrimPrefix(b, "http://")
	// Gemini, Gopher, FTP, Mail are not stripped, to emphasize them, when they are.
	d := strings.TrimSuffix(c, "/")
	return d
}
