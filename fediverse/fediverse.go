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
	"context"
	"strings"

	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"

	"html/template"
	"io"
	"log/slog"
	"net/http"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}

var actorRepo = db.NewActorRepo()

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

func RenderRemoteBookmarks(
	sanitizer wwwports.HTMLSanitizer,
	raws []types.RemoteBookmark,
) (renders []types.RenderedRemoteBookmark) {
	// Gather actor info to prevent duplicate fetches from db
	actors := map[string]*types.Actor{}
	for _, raw := range raws {
		actors[raw.ActorID] = nil
	}
	for actorID := range actors {
		actor, err := actorRepo.GetActorByID(context.Background(), actorID, apports.GetActorsOpts{GetPublicKey: true})
		if err != nil {
			slog.Error("Failed to find actor when gathering remote bookmark actors", "actorID", actorID, "err", err)
			continue // leaves a nil entry, handled below
		}
		actors[actorID] = &actor
	}

	// Rendering
	for _, raw := range raws {
		actor, ok := actors[raw.ActorID]
		if !ok {
			slog.Error("Failed to find actor when rendering remote bookmarks", "actorID", raw.ActorID)
			continue // whatever
		}

		render := types.RenderedRemoteBookmark{
			ID:                  raw.ID,
			AuthorAcct:          actor.Acct(),
			AuthorDisplayedName: actor.PreferredUsername,
			RepostOf:            raw.RepostOf,
			Title:               raw.Title,
			URL:                 raw.URL,
			Tags:                raw.Tags,
		}

		t, err := time.Parse(types.TimeLayout, raw.PublishedAt)
		if err != nil {
			slog.Error("Failed to parse time when rendering remote bookmarks", "err", err)
			continue // whatever
		}
		render.PublishedAt = t

		switch {
		case raw.Source.Valid && raw.SourceType == types.SourcePlainText:
			var (
				escaped        = template.HTMLEscapeString(raw.Source.String)
				taggedNewlines = strings.ReplaceAll(escaped, "\n", "<br/>")
			)
			render.Description = template.HTML("<p>" + taggedNewlines + "</p>")
		case raw.Source.Valid:
			render.Description = myco.MarkupToHTML(raw.Source.String)
		default:
			render.Description = sanitizer.Sanitize(raw.DescriptionHTML)
		}

		renders = append(renders, render)
	}

	return renders
}
