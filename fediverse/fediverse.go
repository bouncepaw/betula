// Package fediverse has some of the Fediverse-related functions.
package fediverse

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing/httpsig"
	"git.sr.ht/~bouncepaw/betula/settings"
	"log"
	"net/http"
	"time"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}

// VerifyRequest returns true if the request is alright. This function makes HTTP requests on your behalf to retrieve the public key.
func VerifyRequest(rq *http.Request, content []byte) bool {
	_, err := httpsig.VerifyRequest(rq, content, func(keyId string) (httpsig.PublicKey, error) {
		pem := db.KeyPemByID(keyId)
		if pem == "" {
			// The zero PublicKey has a None key type, which the underlying VerifyRequest handles well.
			return httpsig.PublicKey{}, nil
		}

		_, pub, err := httpsig.DecodeKey(pem)
		return pub, err
	})
	if err != nil {
		log.Printf("When verifying the signature of request to %s got error: %s\n", rq.URL.RequestURI(), err)
		return false
	}
	return true
}

func OurID() string {
	return settings.SiteURL() + "/@" + settings.AdminUsername()
}
