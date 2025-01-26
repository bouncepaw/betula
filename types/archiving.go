package types

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
)

// Artifact is an artifact in any format stored in database.
type Artifact struct {
	ID            string
	MimeType      sql.NullString
	Data          []byte
	SavedAt       sql.NullString
	LastCheckedAt sql.NullString
}

func (a Artifact) IsCompressed() bool {
	return a.MimeType.Valid && strings.HasSuffix(a.MimeType.String, "+gzip")
}

// NewCompressedDocumentArtifact makes an Artifact from the given
// uncompressed document. Artifact.ID is a base64 representation
// of an SHA-256 hash sum of the document contents. Artifact.MimeType
// is the source MIME type but with "+gzip" added at the end.
// Artifact.Data is gzipped document contents.
//
// Gzip was chosen because it's the most widely accepted content
// compression algorithm in browsers. This way, we can deliver
// the document without intermediary recompression.
func NewCompressedDocumentArtifact(b []byte, mime string) (*Artifact, error) {
	var id string
	{
		var hash = sha256.New().Sum(b)

		var buf strings.Builder
		var encoder = base64.NewEncoder(base64.URLEncoding, &buf)

		var _, err = encoder.Write(hash)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate base64 hash sum: %w", err)
		}
		err = encoder.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to calculate base64 hash sum: %w", err)
		}

		id = buf.String()
	}

	var gzipped []byte
	{
		var buf bytes.Buffer
		var gzipper = gzip.NewWriter(&buf)

		var _, err = gzipper.Write(b)
		if err != nil {
			return nil, fmt.Errorf("failed to compress artifact: %w", err)
		}
		err = gzipper.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to compress artifact: %w", err)
		}

		gzipped = buf.Bytes()
	}

	return &Artifact{
		ID: id,
		MimeType: sql.NullString{
			String: mime + "+gzip",
			Valid:  true,
		},
		Data:          gzipped,
		SavedAt:       sql.NullString{},
		LastCheckedAt: sql.NullString{},
	}, nil
}

type Archive struct {
	ID       int64
	Artifact Artifact
	Note     sql.NullString
}
