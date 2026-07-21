// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package fediverse has some of the Fediverse-related functions.
package fediverse

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/svc/activitypub/parsing"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}

var (
	actorRepo  = db.NewActorRepo()
	noteParser = parsing.NewNoteParser(settings.SiteURL)
)

func PostSignedDocumentToAddress(doc []byte, contentType string, accept string, addr string) ([]byte, int, error) {
	rq, err := http.NewRequest(http.MethodPost, addr, bytes.NewReader(doc))
	if err != nil {
		slog.Error("Failed to prepare signed document request",
			"err", err, "addr", addr)
		return nil, 0, err
	}

	rq.Header.Set("User-Agent", settings.UserAgent())
	rq.Header.Set("Content-Type", contentType)
	rq.Header.Set("Accept", accept)
	signing.SignRequest(rq, doc)

	resp, err := client.Do(rq)
	if err != nil {
		slog.Error("Failed to send signed document request",
			"err", err, "addr", addr)
		return nil, 0, err
	}
	defer resp.Body.Close()

	var (
		bodyReader = io.LimitReader(resp.Body, 1024*1024*10)
		body       []byte
	)
	body, err = io.ReadAll(bodyReader)
	if err != nil {
		slog.Error("Failed to read body", "err", err)
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Non-OK status code returned",
			"err", err, "addr", addr, "status", resp.StatusCode)
	}

	return body, resp.StatusCode, nil
}

func OurID() string {
	return settings.SiteURL() + "/@" + settings.AdminUsername()
}
