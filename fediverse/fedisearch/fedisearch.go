package fedisearch

import (
	"encoding/json"
	"errors"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/types"
	"net/url"
)

type RequestState interface {
	NextPage() *RequestState
}

func NewSearchRequest(q string) {}

type RequestV1 struct {
	Version string `json:"version"`
	Query   string `json:"query"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Actor   string `json:"actor"`
	To      string `json:"to"`
}

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

func NewBetulaProvider(targetActor types.Actor) (*Provider, error) {
	return nil, nil
}

type response struct {
	MoreAvailable int               `json:"more_available"`
	Bookmarks     []activities.Dict `json:"bookmarks"`
}

// State is the current state of a federated search request from
// betulist's point of view.
type State struct {
	query string
	// seen maps actors to number of bookmarks already seen.
	// When requesting more bookmarks, these values become
	// values of the "offset" field.
	seen map[string]int

	// expected maps actors to number of bookmarks expected
	// to be possible to request. These values come from
	// the "moreAvailable" field.
	expected map[string]int

	// notRequestedYet lists actor that have not been
	// requested for bookmarks yet, so it's unknown how many
	// do they have.
	notRequestedYet []string
}

// StateFromFormParams fetches fields with serialized state
// and constructs it from them.
func StateFromFormParams(params url.Values) (*State, error) {
	var (
		s = State{
			query:    params.Get("query"),
			seen:     make(map[string]int),
			expected: make(map[string]int),
		}
		seenJSON     = []byte(params.Get("seen"))
		expectedJSON = []byte(params.Get("expected"))
		err          = errors.Join(
			json.Unmarshal(seenJSON, &s.seen),
			json.Unmarshal(expectedJSON, &s.expected))
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SerializeForForm serializes State into a dictionary that
// is easily used in Go templates. The map's fields are:
// "Query", "Seen", "Expected", "NotRequestedYet". They are
// written in title case to match how we write Go templates.
func (s *State) SerializeForForm() (map[string]string, error) {
	var j = map[string]string{
		"Query": s.query,
	}

	if s.seen != nil {
		var seen, err = json.Marshal(s.seen)
		if err != nil {
			return nil, err
		}
		j["Seen"] = string(seen)
	}

	if s.expected != nil {
		var expected, err = json.Marshal(s.expected)
		if err != nil {
			return nil, err
		}
		j["Expected"] = string(expected)
	}

	if s.notRequestedYet != nil {
		var notRequestedYet, err = json.Marshal(s.notRequestedYet)
		if err != nil {
			return nil, err
		}
		j["NotRequestedYet"] = string(notRequestedYet)
	}

	return j, nil
}

// RequestsToMake returns a list of requests to make.
// It arranges the requests in such a way that no more
// than 65 bookmarks are received after making them.
func (s *State) RequestsToMake() []RequestV1 {

}
