// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package imexsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"

	imexports "git.sr.ht/~bouncepaw/betula/ports/imex"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (svc *Service) Import(
	ctx context.Context,
	params imexports.ImportParams,
	seeker io.ReadSeeker,
) (uint, error) {
	imp, err := svc.pickImporter(seeker)
	if err != nil {
		return 0, err
	}

	bookmarks, err := imp.Import(seeker)
	if err != nil {
		return 0, err
	}

	var (
		tagsToAdd = params.TagsToAdd()
		errs      []error
		okCount   uint
		skipCount uint
	)
	for bm, err := range bookmarks {
		if err != nil {
			slog.Warn("Failed to process bookmark during import", "err", err)
			errs = append(errs, fmt.Errorf("failed to process bookmark: %w", err))
			continue
		}
		if params.MakePublic {
			bm.Visibility = types.Public
		}
		bm.Tags = append(bm.Tags, tagsToAdd...)

		if !params.KeepDuplicate {
			free, err := svc.bookmarkURLIsFree(ctx, bm.URL)
			if err != nil {
				slog.Warn("Failed to check if URL is already bookmarked", "url", bm.URL, "err", err)
				errs = append(errs, fmt.Errorf("failed to check if url %s is bookmarked already: %w", bm.URL, err))
				continue
			}
			if !free {
				slog.Info("Skipping duplicate bookmark", "url", bm.URL)
				skipCount++
				continue
			}
		}

		id, err := svc.bmRepo.InsertBookmark(ctx, bm)
		if err != nil {
			slog.Warn("Failed to insert bookmark", "url", bm.URL, "err", err)
			errs = append(errs, fmt.Errorf("failed to insert bookmark %s: %w", bm.URL, err))
			continue
		}

		slog.Info("Imported bookmark",
			"url", bm.URL, "id", id, "title", bm.Title)
		okCount++
	}

	slog.Info("Bookmark import done",
		"okCount", okCount, "errorCount", len(errs), "skipCount", skipCount)
	return okCount, errors.Join(errs...)
}

func (svc *Service) bookmarkURLIsFree(
	ctx context.Context,
	url string,
) (bool, error) {
	_, err := svc.bmRepo.GetBookmarkIDByURL(ctx, url)
	switch {
	case err == nil: // No error => url is taken.
		return false, nil
	case errors.Is(err, sql.ErrNoRows): // No bookmark found.
		return true, nil
	}
	// Unexpected error.
	return false, err
}

func (svc *Service) pickImporter(seeker io.ReadSeeker) (importer, error) {
	for _, i := range svc.importers {
		ok, err := i.Probe(seeker)
		if err != nil {
			return nil, err
		}
		if ok {
			return i, nil
		}
	}
	return nil, errors.ErrUnsupported
}
