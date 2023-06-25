package search

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"regexp"
)

var excludeTagRe = regexp.MustCompile(`-#([^\s]+)\s*`)
var includeTagRe = regexp.MustCompile(`#([^\s]+)\s*`)

// For searches For the given query.
func For(query string, authorized bool) []types.Post {
	// We extract excluded tags first.
	query, excludedTags := extractWithRegex(query, excludeTagRe)
	query, includedTags := extractWithRegex(query, includeTagRe)

	return db.Search(query, includedTags, excludedTags, authorized)
}

func extractWithRegex(query string, regex *regexp.Regexp) (string, []string) {
	var extracted []string
	for _, result := range regex.FindAllStringSubmatch(query, -1) {
		result := result
		extracted = append(extracted, result[1])
	}
	query = regex.ReplaceAllString(query, "")
	return query, extracted
}
