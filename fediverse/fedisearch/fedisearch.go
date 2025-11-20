// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package fedisearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
	"maps"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"sync"
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
	Query string
	// Seen maps actors to number of bookmarks already seen.
	// When requesting more bookmarks, these values become
	// values of the "offset" field.
	Seen map[string]int

	// Expected maps actors to number of bookmarks expected
	// to be possible to request. These values come from
	// the "moreAvailable" field.
	Expected map[string]int

	// Unseen lists actor that have not been
	// requested for bookmarks yet, so it's unknown how many
	// do they have.
	Unseen []string

	ourID string
}

// StateFromFormParams fetches fields with serialized state
// and constructs it from them.
func StateFromFormParams(params url.Values, ourID string) (*State, error) {
	var (
		s = State{
			Query:    params.Get("query"),
			Seen:     make(map[string]int),
			Unseen:   make([]string, 0),
			Expected: make(map[string]int),
			ourID:    ourID,
		}
		seenJSON     = []byte(params.Get("seen"))
		unseenJSON   = []byte(params.Get("unseen"))
		expectedJSON = []byte(params.Get("expected"))
		err          error
	)
	if len(seenJSON) > 0 {
		err = errors.Join(json.Unmarshal(seenJSON, &s.Seen))
	}
	if len(unseenJSON) > 0 {
		err = errors.Join(json.Unmarshal(unseenJSON, &s.Unseen))
	}
	if len(expectedJSON) > 0 {
		err = errors.Join(json.Unmarshal(expectedJSON, &s.Expected))
	}
	if err != nil {
		return nil, err
	}

	// First page:
	if len(seenJSON) == 0 && len(unseenJSON) == 0 && len(expectedJSON) == 0 {
		var mutuals = db.GetMutuals()
		for _, m := range mutuals {
			s.Unseen = append(s.Unseen, m.ID)
		}
	}
	return &s, nil
}

func (s *State) SeenSerialized() string {
	var bytes, err = json.Marshal(s.Seen)
	if err != nil {
		slog.Error("Error serializing seen", "err", err)
		return ""
	}
	return string(bytes)
}

func (s *State) UnseenSerialized() string {
	var bytes, err = json.Marshal(s.Unseen)
	if err != nil {
		slog.Error("Error serializing unseen", "err", err)
		return ""
	}
	return string(bytes)
}

func (s *State) ExpectedSerialized() string {
	var bytes, err = json.Marshal(s.Expected)
	if err != nil {
		slog.Error("Error serializing expected", "err", err)
		return ""
	}
	return string(bytes)
}

// RequestsToMake returns a list of requests to make.
// It arranges the requests in such a way that about
// 65 bookmarks are expected to be received.
func (s *State) RequestsToMake() []Request {
	var (
		choice   = Choice{}
		requests []Request
	)
	choice.fillFor(maps.Clone(s.Expected), slices.Clone(s.Unseen))

	for actorID, limit := range choice {
		requests = append(requests, Request{
			Version: "v1",
			Query:   s.Query,
			Limit:   limit,
			Offset:  s.Seen[actorID],
			From:    s.ourID,
			To:      actorID,
		})
	}
	return requests
}

func (s *State) FetchPage() ([]types.RenderedRemoteBookmark, *State, error) {
	var newState = &State{
		Query:    s.Query,
		Seen:     maps.Clone(s.Seen),
		Expected: map[string]int{},
		Unseen:   slices.Clone(s.Unseen),
		ourID:    s.ourID,
	}

	var mutex sync.Mutex
	var requestedActors []string
	var bookmarks []types.RemoteBookmark

	var reqs = s.RequestsToMake()
	slog.Info("Making federated search requests",
		"len(reqs)", len(reqs), "query", s.Query)

	var wg sync.WaitGroup
	wg.Add(len(reqs))

	for i, req := range reqs {
		requestedActors = append(requestedActors, req.To)
		go func() {
			defer wg.Done()
			var ok = s.doRequest(i, req, newState, &bookmarks, &mutex)
			if !ok {
				mutex.Lock()
				delete(newState.Expected, req.To)
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	newState.Unseen = slices.DeleteFunc(newState.Unseen, func(s string) bool {
		return slices.Contains(requestedActors, s)
	})
	var rendered = fediverse.RenderRemoteBookmarks(bookmarks)
	return rendered, newState, nil
}

func (s *State) NextPageExpected() bool {
	return len(s.Expected) > 0
}

func (s *State) doRequest(i int, req Request,
	newState *State, bookmarks *[]types.RemoteBookmark,
	mutex *sync.Mutex) (ok bool) {
	var theJSON, err = json.Marshal(req)
	if err != nil {
		slog.Error("Failed to marshal request", "req", req, "i", i, "err", err)
		return false
	}

	theURL, err := url.Parse(req.To)
	if err != nil {
		slog.Error("Failed to parse url", "url", req.To, "err", err)
		return false
	}

	bytes, status, err := fediverse.PostSignedDocumentToAddress(
		theJSON, "application/json", "application/json",
		fmt.Sprintf("https://%s/.well-known/betula-federated-search", theURL.Host))
	if err != nil || status != http.StatusOK {
		slog.Error("Failed to fetch bookmarks",
			"err", err, "status", status, "to", req.To)
		return false
	}

	slog.Info("Sent request",
		"req", req)

	var resp response
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		slog.Error("Failed to unmarshal response", "err", err, "resp", string(bytes), "i", i)
		return false
	}

	slog.Info("Got response", "resp", resp, "to", req.To, "payload", string(bytes))

	var foundBookmarks []types.RemoteBookmark
	for _, bookmark := range resp.Bookmarks {
		bm, err := activities.RemoteBookmarkFromDict(bookmark)
		if err != nil {
			slog.Error("Failed to unmarshal bookmark", "err", err, "bookmark", bm, "i", i)
			continue
		}
		foundBookmarks = append(foundBookmarks, *bm)
	}

	mutex.Lock()
	if resp.MoreAvailable != 0 {
		newState.Expected[req.To] = resp.MoreAvailable
	}
	newState.Seen[req.To] = newState.Seen[req.To] + len(foundBookmarks)
	*bookmarks = append(*bookmarks, foundBookmarks...)
	mutex.Unlock()

	return true
}

type Choice map[string]int

func (choice Choice) fillFor(expected map[string]int, unseen []string) {
	for choice.sum() < 65 {
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
