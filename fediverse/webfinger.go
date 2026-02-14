// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package fediverse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/settings"
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
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Printf("Failed to construct request from ‘%s’\n", requestURL)
		return "", err
	}

	req.Header.Set("User-Agent", settings.UserAgent())
	resp, err := client.Do(req)
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
