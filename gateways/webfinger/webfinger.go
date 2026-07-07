// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package webfingergw

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	webfingerports "git.sr.ht/~bouncepaw/betula/ports/webfinger"
	"git.sr.ht/~bouncepaw/betula/settings"
)

// https://docs.joinmastodon.org/spec/webfinger/

type WebFinger struct {
	client *http.Client
}

var _ webfingerports.WebFinger = &WebFinger{}

func New() *WebFinger {
	return &WebFinger{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (wf *WebFinger) DereferenceAcct(acct webfingerports.Acct) (id string, err error) {
	requestURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=%s", acct.Host, acct.String())
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to construct WebFinger request for %s: %w", acct, err)
	}

	req.Header.Set("User-Agent", settings.UserAgent())
	resp, err := wf.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request WebFinger for %s: %w", acct, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read WebFinger response for %s: %w", acct, err)
	}

	var obj webfingerports.Document
	if err = json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("failed to unmarshal WebFinger document for %s: %w", acct, err)
	}

	return obj.ActivityPubActorID(), nil
}
