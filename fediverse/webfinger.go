package fediverse

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"io"
)

// https://docs.joinmastodon.org/spec/webfinger/

type webfingerDocument struct {
	Links []struct {
		Rel  string `json:"rel"`
		Type string `json:"type"`
		Href string `json:"href"`
	} `json:"links"`
}

func requestIdByWebFingerAcct(user, host string) (id string, err error) {
	requestURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s", host, user, host)

	resp, err := client.Get(requestURL)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	obj := webfingerDocument{}
	if err = json.Unmarshal(data, &obj); err != nil {
		return "", err
	}

	for _, link := range obj.Links {
		if link.Rel == "self" && link.Type == "application/activity+json" && stricks.ValidURL(link.Href) {
			// Found what we were looking for
			return link.Href, nil
		}
	}

	// Mistakes happen
	return "", nil
}
