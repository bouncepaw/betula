package fedisearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/types"
	"maps"
	"math"
	"math/rand"
	"net/url"
	"slices"
)

type Request struct {
	Version string `json:"version"`
	Query   string `json:"query"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	From    string `json:"from"`
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

	// unseen lists actor that have not been
	// requested for bookmarks yet, so it's unknown how many
	// do they have.
	unseen []string

	ourID string
}

func NewSearchRequest(q string, mutuals []string, ourID string) *State {
	return &State{
		query:    q,
		seen:     nil,
		expected: nil,
		unseen:   mutuals,
		ourID:    ourID,
	}
}

// StateFromFormParams fetches fields with serialized state
// and constructs it from them.
func StateFromFormParams(params url.Values, ourID string) (*State, error) {
	var (
		s = State{
			query:    params.Get("query"),
			seen:     make(map[string]int),
			expected: make(map[string]int),
			ourID:    ourID,
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
// "Query", "Seen", "Expected", "Unseen". They are
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

	if s.unseen != nil {
		var unseen, err = json.Marshal(s.unseen)
		if err != nil {
			return nil, err
		}
		j["Unseen"] = string(unseen)
	}

	return j, nil
}

// RequestsToMake returns a list of requests to make.
// It arranges the requests in such a way that about
// 65 bookmarks are expected to be received.
func (s *State) RequestsToMake() []Request {
	var (
		choice   = Choice{}
		requests []Request
	)
	choice.fillFor(maps.Clone(s.expected), slices.Clone(s.unseen))

	for actorID, limit := range choice {
		requests = append(requests, Request{
			Version: "v1",
			Query:   s.query,
			Limit:   limit,
			Offset:  s.seen[actorID],
			From:    s.ourID,
			To:      actorID,
		})
	}
	return requests
}

type Choice map[string]int

func (choice Choice) fillFor(expected map[string]int, unseen []string) {
	for choice.sum() < 65 {
		fmt.Printf("%v\n", expected)
		if len(unseen) > 0 {
			var (
				unfilledSlots = float64(65.0 - choice.sum())
				r             = int(math.Ceil(unfilledSlots / 5))
				picking       = min(r, len(unseen))
			)

			rand.Shuffle(len(unseen), func(i, j int) {
				unseen[i], unseen[j] = unseen[j], unseen[i]
			})
			for i, actor := range unseen {
				if i == picking {
					break
				}

				choice.increaseLimit(actor, 5)
			}
		}

		if len(unseen) == 0 && len(expected) == 0 {
			return
		}

		var (
			unfilledSlots = float64(65.0 - choice.sum())
			r             = int(math.Ceil(unfilledSlots / 5))
			picking       = min(r, len(expected))
			keys          = slices.Collect(maps.Keys(expected))
		)

		rand.Shuffle(len(expected), func(i, j int) {
			keys[i], keys[j] = keys[j], keys[i]
		})
		for i, actor := range keys {
			if i == picking {
				break
			}

			var a = min(5, expected[actor])
			choice.increaseLimit(actor, a)
			if expected[actor] == a {
				delete(expected, actor)
			} else {
				expected[actor] = expected[actor] - a
			}
		}
	}
}

func (choice Choice) sum() int {
	var sum = 0
	for _, v := range choice {
		sum += v
	}
	return sum
}

func (choice Choice) increaseLimit(actorID string, limit int) {
	choice[actorID] = choice[actorID] + limit
}
