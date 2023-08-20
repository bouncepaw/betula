// Package types provides common data types used across all Betula, all conveniently collected in a single box for resolving import cycles nicely.
package types

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strings"

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

type JobCategory string

const (
	NotifyAboutMyRepost JobCategory = "notify about my repost"
	CheckTheirRepost    JobCategory = "check their repost"
)

// Job is a task for Betula to do later.
type Job struct {
	// ID is a unique identifier for the Job. You get it when reading from the database. Do not set it when issuing a new job.
	ID       int
	Category JobCategory
	// Payload is some data.
	Payload []byte
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

// not really a type:

const TimeLayout = "2006-01-02 15:04:05"

func StripCommonProtocol(a string) string {
	b := strings.TrimPrefix(a, "https://")
	c := strings.TrimPrefix(b, "http://")
	// Gemini, Gopher, FTP, Mail are not stripped, to emphasize them, when they are.
	d := strings.TrimSuffix(c, "/")
	return d
}

const (
	PostsPerPage uint = 64
)
