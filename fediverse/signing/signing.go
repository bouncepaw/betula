// Package sign manages HTTP signatures and storing a pair of private and public keys. This package is a wrapper around Ted of the Honk's httpsig package, which is vendored in Betula.
package signing

import (
	"crypto/rand"
	"crypto/rsa"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing/httpsig"
	"git.sr.ht/~bouncepaw/betula/settings"
	"log"
	"net/http"
)

// SignRequest signs the request.
func SignRequest(rq *http.Request, content []byte) {
	keyId := settings.SiteURL() + "#main-key"
	httpsig.SignRequest(keyId, privateKey, rq, content)
}

// VerifyRequest returns true if the request is alright. This function makes HTTP requests on your behalf to retrieve the public key.
func VerifyRequest(rq *http.Request, content []byte) bool {
	_, err := httpsig.VerifyRequest(rq, content, func(keyId string) (httpsig.PublicKey, error) {
		pem := db.GetPublicKeyPEM(keyId)
		if pem == "" {
			// The zero PublicKey has a None key type, which the underlying VerifyRequest handles well.
			return httpsig.PublicKey{}, nil
		}

		_, pub, err := httpsig.DecodeKey(pem)
		return pub, err
	})
	if err != nil {
		log.Printf("When verifying the signature of request to %s from %s got error: %s\n", rq.URL.RequestURI(), rq.Host, err)
		return false
	}
	return true
}

var (
	privateKey   httpsig.PrivateKey
	publicKey    httpsig.PublicKey
	publicKeyPEM string
)

func PublicKey() string {
	return publicKeyPEM
}

func setKeys(privateKeyPEM string) {
	var err error
	privateKey, publicKey, err = httpsig.DecodeKey(privateKeyPEM)
	if err != nil {
		log.Fatalf("When decoding private key PEM: %s\n", err)
	}

	publicKeyPEM, err = httpsig.EncodeKey(publicKey.Key)
	if err != nil {
		log.Fatalf("When encoding public key PEM: %s\n", err)
	}
}

// EnsureKeysFromDatabase reads the keys from the database and remembers them. If they are not found, it comes up with new ones and saves them. This function might crash the application.
func EnsureKeysFromDatabase() {
	privKeyPEM := db.MetaEntry[string](db.BetulaMetaPrivateKey)

	// No key found? Make a new one and write it to DB.
	if privKeyPEM == "" {
		log.Println("Generating a new pair of RSA keys")
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("When generating new keys: %s\n", err)
		}

		privKeyPEM, err = httpsig.EncodeKey(priv)
		if err != nil {
			log.Fatalf("When generating private key PEM: %s\n", err)
		}

		db.SetMetaEntry(db.BetulaMetaPrivateKey, privKeyPEM)
	}

	setKeys(privKeyPEM)
}