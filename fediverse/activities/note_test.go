// SPDX-FileCopyrightText: 2022-2025 Betula contributors
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"embed"
	"io"
	"testing"

	"github.com/nalgeon/be"
)

//go:embed testdata/*
var fs embed.FS

func TestGuessCreateNote(t *testing.T) {
	f, err := fs.Open("testdata/Create{Note} 1.json")
	if err != nil {
		panic(err)
	}

	raw, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	report, err := Guess(raw)
	be.Err(t, err, nil)
	r, ok := report.(CreateNoteReport)
	be.True(t, ok)
	be.Equal(t, len(r.Bookmark.Tags), 1)
}
