package fedisearch

import (
	"encoding/json"
	"errors"
	"git.sr.ht/~bouncepaw/betula/fediverse"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported version")
	ErrWrongTo            = errors.New("field to does not match")
	ErrNotMutual          = errors.New("not mutual")
)

func ParseAPIRequest(bytes []byte) (*Request, error) {
	var req Request
	var err = json.Unmarshal(bytes, &req)
	if err != nil {
		return nil, err
	}

	fromActor, err := fediverse.RequestActorByID(req.From)
	if err != nil {
		return nil, err
	}

	switch {
	case req.Version != "v1":
		return nil, ErrUnsupportedVersion
	case req.To != fediverse.OurID():
		return nil, ErrWrongTo
	case !(fromActor.SubscriptionStatus.WeFollowThem() && fromActor.SubscriptionStatus.TheyFollowUs()):
		return nil, ErrNotMutual
	}

	return &req, nil
}
