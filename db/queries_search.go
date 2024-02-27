package db

import (
	"sort"
	"strings"

	"git.sr.ht/~bouncepaw/betula/types"
)

func Search(text string, includedTags []string, excludedTags []string, repostsOnly, authorized bool, page uint) (posts []types.Bookmark, totalPosts uint) {
	text = strings.ToLower(text)
	sort.Strings(includedTags)
	sort.Strings(excludedTags)

	const q = `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf
from Posts
where DeletionTime is null and (Visibility = 1 or ?)
order by CreationTime desc
`
	rows := mustQuery(q, authorized)

	var unfilteredPosts []types.Bookmark
	for rows.Next() {
		var post types.Bookmark
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		unfilteredPosts = append(unfilteredPosts, post)
	}

	var i uint = 0
	var ignoredPosts uint = 0
	postsToBeIgnored := (page - 1) * types.PostsPerPage

	// ‘Say, Bouncepaw, why did not you implement tag inclusion/exclusion
	//  part in SQL directly?’, some may ask.
	// ‘I did, and it was not worth it’, so I would respond.
	//
	// Addendum: I tried to make case-insensitive search in SQL too, and
	// failed loudly. Now all the search is done in Go. Per aspera ad
	// astra.
	//
	// We can't even parallelize it.
	for _, post := range unfilteredPosts {
		if !textOK(post, text) {
			continue
		}

		post.Tags = TagsForPost(post.ID)
		if !tagsOK(post.Tags, includedTags, excludedTags) {
			continue
		}

		isRepost := post.RepostOf != nil
		if !isRepost && repostsOnly {
			continue
		}

		totalPosts++
		if ignoredPosts >= postsToBeIgnored && i < types.PostsPerPage {
			posts = append(posts, post)
			i++
		} else {
			ignoredPosts++
		}
	}
	return posts, totalPosts
}

// true if keep, false if discard
func textOK(post types.Bookmark, text string) bool {
	return strings.Contains(strings.ToLower(post.Title), text) ||
		strings.Contains(strings.ToLower(post.Description), text) ||
		strings.Contains(strings.ToLower(post.URL), text)
}

// true if keep, false if discard. All slices are sorted.
func tagsOK(postTags []types.Tag, includedTags, excludedTags []string) bool {
	J, K := len(includedTags), len(excludedTags)
	j, k := 0, 0
	includeMask := make([]bool, J)
	for _, postTag := range postTags {
		name := postTag.Name
		switch {
		case k < K && excludedTags[k] == name:
			return false
		case j < J && includedTags[j] == name:
			includeMask[j] = true
			j++
			continue
		}

		for j < J && includedTags[j] < name {
			j++
		}

		for k < K && excludedTags[k] < name {
			k++
		}
	}

	for _, marker := range includeMask {
		if marker == false {
			return false
		}
	}
	return true
}
