// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package archivingsvc

import (
	archivingports "git.sr.ht/~bouncepaw/betula/ports/archiving"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
)

type Service struct {
	fetcher      archivingports.Fetcher
	archivesRepo archivingports.ArchivesRepo
}

var _ archivingports.Service = &Service{}

func New(
	fetcher archivingports.Fetcher,
	archivesRepo archivingports.ArchivesRepo,
) *Service {
	return &Service{
		fetcher:      fetcher,
		archivesRepo: archivesRepo,
	}
}

func (svc *Service) Archive(bookmark types.Bookmark) (int64, error) {
	var bytes, mime, err = svc.fetcher.Fetch(bookmark.URL)
	if err != nil {
		slog.Error("Obelisk failed to fetch an archive of the page",
			"url", bookmark.URL, "err", err)
		return 0, err
	}

	artifact, err := types.NewCompressedDocumentArtifact(bytes, mime)
	if err != nil {
		slog.Error("Failed to compress the new archive",
			"url", bookmark.URL, "err", err)
		return 0, err
	}

	archiveID, err := svc.archivesRepo.Store(int64(bookmark.ID), artifact)
	if err != nil {
		slog.Error("Failed to store the new archive",
			"url", bookmark.URL, "err", err)
		return 0, err
	}

	return archiveID, nil
}
