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
	}
	report, ok := report.(CreateNoteReport)
	if !ok {
		t.Error("wrong type")
	}
}
