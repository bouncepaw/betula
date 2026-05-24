// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package imexsvc

import (
	"context"
	"fmt"
	"io"
	"iter"
	"time"

	imexports "git.sr.ht/~bouncepaw/betula/ports/imex"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (svc *Service) Export(
	ctx context.Context,
	params imexports.ExportParams,
	w io.Writer,
) error {
	exp, err := svc.pickExporter(params)
	if err != nil {
		return err
	}

	var (
		iterErr   error
		bookmarks = svc.allBookmarks(ctx, params.IncludePrivate, &iterErr)
	)
	if err = exp.Export(bookmarks, w, time.Now()); err != nil {
		return fmt.Errorf("failed to export bookmarks: %w", err)
	}
	return iterErr
}

func (svc *Service) allBookmarks(
	ctx context.Context,
	includePrivate bool,
	errOut *error,
) iter.Seq[types.Bookmark] {
	return func(yield func(types.Bookmark) bool) {
		for page := uint(1); ; page++ {
			bookmarks, total, err := svc.bmRepo.Bookmarks(ctx, includePrivate, page)
			if err != nil {
				*errOut = err
				return
			}
			for _, bm := range bookmarks {
				if !yield(bm) {
					return
				}
			}
			if uint(len(bookmarks)) == 0 || total <= types.BookmarksPerPage*page {
				return
			}
		}
	}
}

func (svc *Service) pickExporter(params imexports.ExportParams) (exporter, error) {
	exp, ok := svc.exporters[params.Format]
	if !ok || exp == nil {
		return nil, fmt.Errorf("no exporter found for format %q", params.Format)
	}
	return exp, nil
}
