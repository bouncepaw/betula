package readpage

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"io"
	"time"
)

// https://docs.joinmastodon.org/spec/webfinger/

func GetWebFinger(user, host string) (wa types.WebfingerAcct, found bool, err error) {
	requestURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s", host, user, host)

	resp, err := client.Get(requestURL)
	if err != nil {
		return wa, false, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return wa, false, err
	}

	obj := map[string]any{}
	if err = json.Unmarshal(data, &obj); err != nil {
		return wa, false, err
	}

	links, ok := obj["links"].([]any)
	if !ok {
		return wa, false, errors.New("links field not an array")
	}

	for _, linkUntyped := range links {
		linkUntyped, ok := linkUntyped.(map[string]any)
		if !ok {
			return wa, false, errors.New("link is not an object")
		}
		rel, ok1 := linkUntyped["rel"].(string)
		typ, ok2 := linkUntyped["type"].(string)
		href, ok3 := linkUntyped["href"].(string)
		if !ok1 || !ok2 || !ok3 {
			return wa, false, errors.New("a field in link is not a string")
		}
		if rel == "self" && typ == "application/activity+json" && stricks.ValidURL(href) {
			// Found what we were looking for
			return types.WebfingerAcct{
				Acct:          user + "@" + host,
				ActorURL:      href,
				Document:      data,
				LastCheckedAt: time.Now().Format(types.TimeLayout),
			}, true, nil
		}
	}

	// Mistakes happen
	return wa, false, nil
}
