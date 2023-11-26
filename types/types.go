// Package types provides common data types used across all Betula, all conveniently collected in a single box for resolving import cycles nicely.
package types

import (
	"fmt"
	"golang.org/x/net/idna"
	"html/template"
	"math"
	"net/url"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/mycomarkup/v5/util"
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
	// RepostOf is URL of the post reposted. Nil if this is an original post.
	RepostOf *string
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

func TagsFromStringSlice(ss []string) []Tag {
	tags := make([]Tag, len(ss))
	for i, tag := range ss {
		tags[i] = Tag{Name: tag}
	}
	return tags
}

type Settings struct {
	NetworkHost string
	NetworkPort uint
	// SiteName is a plaintext name of the site.
	SiteName string
	// SiteTitle is a hypertext title shown in the top left corner, in a <h1>.
	SiteTitle                 template.HTML
	SiteDescriptionMycomarkup string
	SiteURL                   string
	CustomCSS                 string
	FederationEnabled         bool
}

type Page struct {
	Number    uint
	URL       string
	IsCurrent bool
	IsPrev    bool
	IsNext    bool
}

func countPages(totalPosts uint) uint {
	return uint(math.Ceil(float64(totalPosts) / float64(PostsPerPage)))
}

func PaginatorFromURL(url *url.URL, currentPage uint, totalPosts uint) (pages []Page) {
	totalPages := countPages(totalPosts)
	values := url.Query()
	pages = make([]Page, totalPages)

	var i uint = 0
	for ; i < totalPages; i++ {
		page := i + 1
		values.Set("page", fmt.Sprintf("%d", page))
		url.RawQuery = values.Encode()

		pages[i] = Page{
			Number:    page,
			URL:       url.String(),
			IsCurrent: currentPage == page,
			IsPrev:    (currentPage - 1) == page,
			IsNext:    (currentPage + 1) == page,
		}
	}

	return pages
}

type RepostInfo struct {
	Timestamp time.Time
	URL       string
	Name      string
}

// TimeLayout is the time layout used across Betula.
const TimeLayout = "2006-01-02 15:04:05"

// CleanerLinkParts returns the link a with https:// or http:// prefix and the / suffix,
// percent-encoding reversed and Punycode decoded.
//
// Link is returned in two parts: scheme + authority and the rest (path, query, fragment).
func CleanerLinkParts(a string) (string, string) {
	u, err := url.Parse(a)
	if err != nil {
		// Welp, we tried our best.
		return a, ""
	}

	var hostPart string
	if u.Scheme != "http" && u.Scheme != "https" {
		// Gemini, Gopher, FTP, Mail etc are not stripped to emphasize them.
		hostPart += fmt.Sprintf("%s:", u.Scheme)
		// "Opaque" is defined for schemes like `mailto:` or tel:`, where there is no `//`.
		if u.Opaque == "" {
			hostPart += "//"
		}
	}

	if u.User != nil {
		hostPart += u.User.String()
	}

	if u.Host != "" {
		host, err := idna.ToUnicode(u.Host)
		if err != nil {
			// Was worth a shot.
			host = u.Host
		}
		hostPart += host
	}

	if u.Opaque != "" {
		hostPart += u.Opaque
	}

	pathPart := ""

	path := strings.TrimSuffix(u.Path, "/")
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			pathPart += "/"
		}
		pathPart += path
	}

	if u.RawQuery != "" {
		query, err := url.QueryUnescape(u.RawQuery)
		if err != nil {
			// Better luck next time.
			query = u.RawQuery
		}

		pathPart += "?" + query
	}

	if u.Fragment != "" {
		pathPart += "#" + u.Fragment
	}

	return hostPart, pathPart
}

// Same as CleanerLinkParts, but merges the parts back into one url.
func CleanerLink(a string) string {
	left, right := CleanerLinkParts(a)
	return left + right
}

const (
	PostsPerPage uint = 64
)
