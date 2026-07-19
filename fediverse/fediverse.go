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
	"database/sql"
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
	"git.sr.ht/~bouncepaw/betula/svc/activitypub/parsing"
	"git.sr.ht/~bouncepaw/betula/types"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}

var (
	actorRepo          = db.NewActorRepo()
	remoteBookmarkRepo = db.NewRemoteBookmarkRepo()
	noteParser         = parsing.NewNoteParser()
)

func renderRemoteDescription(
	sanitizer wwwports.HTMLSanitizer,
	source sql.NullString,
	sourceType types.SourceType,
	descriptionHTML template.HTML,
) template.HTML {
	switch {
	case source.Valid && sourceType == types.SourcePlainText:
		var (
			escaped        = template.HTMLEscapeString(source.String)
			taggedNewlines = strings.ReplaceAll(escaped, "\n", "<br/>")
		)
		return template.HTML("<p>" + taggedNewlines + "</p>")
	case source.Valid:
		return myco.MarkupToHTML(source.String)
	default:
		return sanitizer.Sanitize(descriptionHTML)
	}
}

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

		content := raw
		var originalAuthorAcct, originalAuthorName, originalWebURL string
		if raw.RemarkedID.Valid {
			original, ok := remoteBookmarkRepo.GetRemoteBookmarkByID(raw.RemarkedID.String)
			if !ok {
				// TODO: dereference the original when we don't have it locally.
				slog.Warn("Skipping remark: remarked original not found",
					"remarkID", raw.ID, "remarkedID", raw.RemarkedID.String)
				continue
			}
			originalWebURL = original.RepresentationURL()
			content = original

			if origActor, err := actorRepo.GetActorByID(context.Background(), original.ActorID, apports.GetActorsOpts{}); err == nil {
				originalAuthorAcct = origActor.Acct()
				originalAuthorName = origActor.PreferredUsername
			} else {
				slog.Warn("Failed to find original author actor when rendering remark",
					"remarkID", raw.ID, "actorID", original.ActorID, "err", err)
			}
		}

		render := types.RenderedRemoteBookmark{
			ID:                          raw.ID,
			AuthorAcct:                  actor.Acct(),
			AuthorDisplayedName:         actor.PreferredUsername,
			RemarkedID:                  raw.RemarkedID,
			OriginalAuthorAcct:          originalAuthorAcct,
			OriginalAuthorDisplayedName: originalAuthorName,
			OriginalWebURL:              originalWebURL,
			Title:                       content.Title,
			URL:                         content.URL,
			WebURL:                      raw.RepresentationURL(),
			Tags:                        content.Tags,
		}

		t, err := time.Parse(types.TimeLayout, raw.PublishedAt)
		if err != nil {
			slog.Error("Failed to parse time when rendering remote bookmarks", "err", err)
			continue // whatever
		}
		render.PublishedAt = t

		render.Description = renderRemoteDescription(sanitizer, content.Source, content.SourceType, content.DescriptionHTML)

		renders = append(renders, render)
	}

	return renders
}
