// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"embed"
	"io"
	"testing"
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
	if err != nil {
		t.Error(err)
		return
	}
	r, ok := report.(CreateNoteReport)
	if !ok {
		t.Error("wrong type")
	}

	if len(r.Bookmark.Tags) != 1 {
		t.Error("tag len mismatch")
	}
}
