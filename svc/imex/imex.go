// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package imexsvc

import (
	"io"
	"iter"
	"time"

	imexports "git.sr.ht/~bouncepaw/betula/ports/imex"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/svc/imex/internal/exporters"
	"git.sr.ht/~bouncepaw/betula/svc/imex/internal/importers"
	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	Service struct {
		importers []importer
		exporters map[imexports.ExportFormat]exporter
		bmRepo    likingports.LocalBookmarkRepository
	}
	importer interface {
		Probe(io.ReadSeeker) (bool, error)
		Import(io.Reader) (iter.Seq2[types.Bookmark, error], error)
	}
	exporter interface {
		Export(iter.Seq[types.Bookmark], io.Writer, time.Time) error
	}
)

var _ imexports.Service = &Service{}

func New(
	bmRepo likingports.LocalBookmarkRepository,
	siteNameFn func() string,
) *Service {
	return &Service{
		bmRepo: bmRepo,
		importers: []importer{
			importers.NewNetscapeImporter(),
		},
		exporters: map[imexports.ExportFormat]exporter{
			imexports.ExportFormatNetscape: exporters.NewNetscapeExporter(siteNameFn),
			imexports.ExportFormatPinboard: nil,
			imexports.ExportFormatCSV:      nil,
		},
	}
}
