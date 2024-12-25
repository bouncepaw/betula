package fedisearch

import "git.sr.ht/~bouncepaw/betula/types"

// Provider provides search results as per the query.
type Provider interface {
	// QueryV1 requests search results from the provider. The query
	// parsing procedure is up to the provider. Future versions might
	// have different API.
	//
	// Cursor is optional, and is an ActivityPub ID of the oldest
	// bookmark on the previous page. A nil cursor means we are
	// asking for the first page.
	QueryV1(query string, cursor *string) ([]types.RenderedRemoteBookmark, error)
}
