// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package types

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// Artifact is an artifact in any format stored in database.
type Artifact struct {
	ID        string
	MimeType  string
	Data      []byte
	IsGzipped bool
	Size      int
}

// NewCompressedDocumentArtifact makes an Artifact from the given
// uncompressed document. Artifact.ID is a base64 representation
// of an SHA-256 hash sum of the document contents. Artifact.MimeType
// is the source MIME type. Artifact.IsGzipped is true.
// Artifact.Data is gzipped document contents.
//
// Gzip was chosen because it's the most widely accepted content
// compression algorithm in browsers. This way, we can deliver
// the document without intermediary recompression.
func NewCompressedDocumentArtifact(b []byte, mime string) (*Artifact, error) {
	var id string
	{
		var hash = sha256.New()
		var _, err = hash.Write(b)
		if err != nil {
			return nil, fmt.Errorf("failed to write bytes to sha256: %w", err)
		}

		var buf strings.Builder
		var encoder = base64.NewEncoder(base64.RawURLEncoding, &buf)

		_, err = encoder.Write(hash.Sum(nil))
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
		ID:        id,
		MimeType:  mime,
		Data:      gzipped,
		IsGzipped: true,
	}, nil
}

func (a *Artifact) HumanSize() string {
	switch {
	case a.Size == 0:
		return "empty"
	case a.Size < 1024:
		return fmt.Sprintf("%d B", a.Size)
	case a.Size < 1024*1024:
		return fmt.Sprintf("%.2f KiB", float64(a.Size)/float64(1024))
	default:
		return fmt.Sprintf("%.2f MiB", float64(a.Size)/float64(1024*1024))
	}
}

var reMime = regexp.MustCompile(`[a-z]+/([a-z]+).*`)

func (a *Artifact) HumanMimeType() string {
	matches := reMime.FindStringSubmatch(a.MimeType)
	if len(matches) != 2 {
		return a.MimeType
	}
	return matches[1]
}

type Archive struct {
	ID       int64
	Artifact Artifact
	SavedAt  sql.NullString
}
