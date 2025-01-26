package archiving

import (
	"codeberg.org/bouncepaw/obelisk-ng"
	"context"
	"time"
)

// Archiver archives documents.
type Archiver interface {
	// Fetch fetches an archive copy for the document identified by URL.
	// Returns contents, MIME-type and a possible error.
	Fetch(url string) ([]byte, string, error)
}

// ObeliskArchiver fetched archive copies using
// the obelisk-ng library.
type ObeliskArchiver struct {
	*obelisk.Archiver
}

func NewObeliskArchiver() *ObeliskArchiver {
	var a = ObeliskArchiver{
		Archiver: &obelisk.Archiver{
			Cache:            nil,
			EnableLog:        true,
			EnableVerboseLog: false,
			DisableJS:        false,
			DisableCSS:       false,
			DisableEmbeds:    false,
			DisableMedias:    false,
			RequestTimeout:   time.Second * 15,
			MaxRetries:       3,
		},
	}
	a.Archiver.Validate()
	return &a
}

func (o *ObeliskArchiver) Fetch(url string) ([]byte, string, error) {
	return o.Archiver.Archive(context.Background(), obelisk.Request{
		URL: url,
	})
}
