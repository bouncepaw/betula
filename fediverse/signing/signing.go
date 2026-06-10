// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package signing manages HTTP signatures and managing a pair of private and public keys. This package is a wrapper around Ted of the Honk's httpsig package.
package signing

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"net/http"
	"os"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/pkg/httpsig"
	"git.sr.ht/~bouncepaw/betula/ports/settings"
	"git.sr.ht/~bouncepaw/betula/settings"
)

var settingsRepo = &db.SettingsRepo{}
var actorRepo = db.NewActorRepo()

// SignRequest signs the request.
func SignRequest(rq *http.Request, content []byte) {
	keyId := settings.SiteURL() + "/@" + settings.AdminUsername() + "#main-key"
	httpsig.SignRequest(keyId, privateKey, rq, content)
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
		slog.Error("Failed to decode private key PEM", "err", err)
		os.Exit(1)
	}

	publicKeyPEM, err = httpsig.EncodeKey(publicKey.Key)
	if err != nil {
		slog.Error("Failed to encode public key PEM", "err", err)
		os.Exit(1)
	}
}

// EnsureKeysFromDatabase reads the keys from the database and remembers them. If they are not found, it comes up with new ones and saves them. This function might crash the application.
func EnsureKeysFromDatabase() {
	ctx := context.Background()
	var pem string
	privKeyPEMMaybe, err := settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaPrivateKey)
	if err != nil {
		slog.Error("Failed to read private key PEM", "err", err)
		os.Exit(1)
	}
	if !privKeyPEMMaybe.Valid || privKeyPEMMaybe.String == "" {
		slog.Info("Generating a new pair of RSA keys")
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			slog.Error("Failed to generate RSA keys", "err", err)
			os.Exit(1)
		}

		pem, err = httpsig.EncodeKey(priv)
		if err != nil {
			slog.Error("Failed to encode private key PEM", "err", err)
			os.Exit(1)
		}

		if err := settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaPrivateKey, pem); err != nil {
			slog.Error("Failed to save private key PEM", "err", err)
			os.Exit(1)
		}
		setKeys(pem)
	} else {
		setKeys(privKeyPEMMaybe.String)
	}
}

// VerifyRequestSignature returns true if the request has correct signature. This function makes HTTP requests on your behalf to retrieve the public key.
func VerifyRequestSignature(rq *http.Request, content []byte) bool {
	_, err := httpsig.VerifyRequest(rq, content, func(keyID string) (httpsig.PublicKey, error) {
		pem, err := actorRepo.KeyPemByID(context.Background(), keyID)
		if err != nil {
			return httpsig.PublicKey{}, err
		}
		if pem == "" {
			// The zero PublicKey has a None key type, which the underlying VerifyRequest handles well.
			return httpsig.PublicKey{}, nil
		}

		_, pub, err := httpsig.DecodeKey(pem)
		return pub, err
	})
	if err != nil {
		slog.Error("Failed to verify request signature", "uri", rq.URL.RequestURI(), "err", err)
		return false
	}
	return true
}
