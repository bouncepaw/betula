// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package imexports

import (
	"context"
	"io"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	Service interface {
		// Import returns errors.ErrUnsupported if no importer supports the format.
		Import(context.Context, ImportParams, io.ReadSeeker) (uint, error)
		// Export returns
		Export(context.Context, ExportParams, io.Writer) error
	}

	ImportParams struct {
		AddTags       []string
		KeepDuplicate bool
		MakePublic    bool
	}
	ExportParams struct {
		IncludePrivate bool
		Format         ExportFormat
	}
	ExportFormat string
)

const (
	ExportFormatNetscape ExportFormat = "netscape"
	ExportFormatPinboard ExportFormat = "pinboard"
	ExportFormatCSV      ExportFormat = "csv"
)

func (ip ImportParams) TagsToAdd() []types.Tag {
	var tagsToAdd []types.Tag
	for _, tag := range ip.AddTags {
		tagsToAdd = append(tagsToAdd, types.Tag{Name: tag})
	}
	return tagsToAdd
}
