// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remotebookmarkssvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (r *Service) ourID() string {
	return r.siteURLFn() + "/@" + r.adminUsernameFn()
}

func (r *Service) renderRemoteDescription(
	source sql.NullString,
	sourceType types.SourceType,
	descriptionHTML template.HTML,
) template.HTML {
	// NOTE(bouncepaw): Not the prettiest solution.
	return RenderRemoteDescription(r.sanitizer, source, sourceType, descriptionHTML)
}

func RenderRemoteDescription(
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

func (r *Service) remarkedRemoteBookmarksFor(
	ctx context.Context,
	rbs []types.RemoteBookmark,
) map[string]types.RemoteBookmark {
	remarked := map[string]types.RemoteBookmark{}

	for _, rb := range rbs {
		if !rb.IsRemark() {
			continue
		}
		remarkedID := rb.RemarkedID.String

		if _, err := strconv.Atoi(remarkedID); err == nil {
			continue
		}
		if _, ok := remarked[remarkedID]; ok {
			continue
		}

		original, err := r.GetRemoteBookmarkByID(ctx, remarkedID)
		if err != nil {
			slog.Error("Failed to fetch remarked original",
				"err", err, "remarkID", rb.ID, "remarkedID", remarkedID)
			continue
		}

		remarked[remarkedID] = original
	}

	return remarked
}

func (r *Service) GetRemoteBookmarkByID(
	ctx context.Context,
	id string,
) (types.RemoteBookmark, error) {
	bookmark, ok := r.remoteBookmarkRepo.GetRemoteBookmarkByID(id)
	if ok {
		return bookmark, nil
	}

	slog.Info("Remote bookmark not found in db, trying to dereference", "id", id)

	bookmark, err := r.activityPub.DerefRemoteBookmark(ctx, id)
	if err != nil {
		return types.RemoteBookmark{}, fmt.Errorf("failed to dereference remote bookmark %s: %w", id, err)
	}

	r.remoteBookmarkRepo.InsertRemoteBookmark(bookmark)
	return bookmark, nil
}

func (r *Service) authorsFor(
	ctx context.Context,
	rbs []types.RemoteBookmark,
) map[string]apports.Actor {
	authors := map[string]apports.Actor{}

	for _, rb := range rbs {
		if _, ok := authors[rb.ActorID]; ok {
			continue
		}

		if rb.ActorID == r.ourID() {
			authors[rb.ActorID] = ownActor{
				id:       rb.ActorID,
				username: r.adminUsernameFn(),
				domain:   r.siteDomainFn(),
			}
			continue
		}

		actor, err := r.activityPub.ActorByID(ctx, rb.ActorID, apports.GetActorsOpts{GetPublicKey: true})
		if err != nil {
			slog.Error("Failed to find actor when gathering remote bookmark authors",
				"actorID", rb.ActorID, "err", err)
			continue
		}
		authors[rb.ActorID] = actor
	}

	return authors
}

func (r *Service) fillLocalRemark(
	ctx context.Context,
	render *types.RenderedRemoteBookmark,
	localID int,
) bool {
	localBookmark, err := r.localBookmarkRepo.GetBookmarkByID(ctx, localID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Warn("Skipping remark of missing local bookmark",
			"localID", localID, "remarkID", render.ID, "remarkedID", render.RemarkedID.String)
		return false
	}
	if err != nil {
		slog.Error("Failed to find bookmark when rendering remote bookmark",
			"localID", localID, "remarkID", render.ID, "remarkedID", render.RemarkedID.String, "err", err)
		return false
	}
	render.Title = localBookmark.Title
	render.URL = localBookmark.URL
	render.OriginalAuthorDisplayedName = r.adminUsernameFn()
	render.OriginalAuthorAcct = fmt.Sprintf("@%s@%s", r.adminUsernameFn(), r.siteDomainFn())
	render.OriginalWebURL = fmt.Sprintf("/%d", localID)
	render.Description = r.renderRemoteDescription(
		bxstr.NullStringFromString(localBookmark.Description),
		types.SourceMycomarkup,
		"",
	)
	return true
}

func (r *Service) fillRemoteRemark(
	render *types.RenderedRemoteBookmark,
	raw types.RemoteBookmark,
	remarked map[string]types.RemoteBookmark,
	authors map[string]apports.Actor,
) bool {
	original, ok := remarked[raw.RemarkedID.String]
	if !ok {
		slog.Warn("Skipping remark whose original could not be resolved",
			"remarkID", raw.ID, "remarkedID", raw.RemarkedID.String)
		return false
	}
	render.Title = original.Title
	render.URL = original.URL
	render.OriginalWebURL = original.RepresentationURL()
	render.Description = r.renderRemoteDescription(original.Source, original.SourceType, original.DescriptionHTML)

	if origActor, ok := authors[original.ActorID]; ok {
		render.OriginalAuthorAcct = origActor.Acct()
		render.OriginalAuthorDisplayedName = origActor.PreferredUsername()
	} else {
		slog.Warn("Failed to find original author actor when rendering remark",
			"remarkID", raw.ID, "actorID", original.ActorID)
	}
	return true
}

func (r *Service) Render(
	ctx context.Context,
	raws []types.RemoteBookmark,
) ([]types.RenderedRemoteBookmark, error) {
	remarked := r.remarkedRemoteBookmarksFor(ctx, raws)

	authored := slices.Concat(raws, slices.Collect(maps.Values(remarked)))
	authors := r.authorsFor(ctx, authored)

	var renders []types.RenderedRemoteBookmark
	for _, raw := range raws {
		actor, ok := authors[raw.ActorID]
		if !ok {
			slog.Error("Failed to find actor when rendering remote bookmarks", "actorID", raw.ActorID)
			continue
		}

		render := types.RenderedRemoteBookmark{
			ID:                  raw.ID,
			AuthorAcct:          actor.Acct(),
			AuthorDisplayedName: actor.PreferredUsername(),
			RemarkedID:          raw.RemarkedID,
			WebURL:              raw.RepresentationURL(),
			Tags:                raw.Tags,
		}

		switch {
		case raw.IsRemark():
			if raw.HasRemarkText() {
				render.RemarkText = r.renderRemoteDescription(raw.Source, raw.SourceType, raw.DescriptionHTML)
			}
			if id, err := strconv.Atoi(raw.RemarkedID.String); err == nil {
				if !r.fillLocalRemark(ctx, &render, id) {
					continue
				}
			} else if !r.fillRemoteRemark(&render, raw, remarked, authors) {
				continue
			}
		case raw.IsRegularBookmark():
			render.Title = raw.Title
			render.URL = raw.URL
			render.Description = r.renderRemoteDescription(raw.Source, raw.SourceType, raw.DescriptionHTML)
		default:
			continue
		}

		t, err := time.Parse(types.TimeLayout, raw.PublishedAt)
		if err != nil {
			slog.Error("Failed to parse time when rendering remote bookmarks", "err", err)
			continue
		}
		render.PublishedAt = t

		renders = append(renders, render)
	}

	return renders, nil
}
