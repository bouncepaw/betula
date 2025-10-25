package archivingsvc

import (
	"codeberg.org/bouncepaw/obelisk-ng"
	"context"
	"time"
)

// ObeliskFetcher fetches archive copies using
// the obelisk-ng library.
type ObeliskFetcher struct {
	*obelisk.Archiver
}

func NewObeliskFetcher() *ObeliskFetcher {
	var a = ObeliskFetcher{
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

func (o *ObeliskFetcher) Fetch(url string) ([]byte, string, error) {
	return o.Archiver.Archive(context.Background(), obelisk.Request{
		URL: url,
	})
}
