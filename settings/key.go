package settings

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"git.sr.ht/~bouncepaw/betula/db"
	"log"
)

var (
	privPEM []byte
	pubPEM  []byte
)

func PrivateKey() []byte { return privPEM }
func PublicKey() []byte  { return pubPEM }

func ensureKeys() {
	privPEM = db.MetaEntry[[]byte](db.BetulaMetaPrivateKey)
	pubPEM = db.MetaEntry[[]byte](db.BetulaMetaPublicKey)
	if privPEM == nil || pubPEM == nil {
		generateKeys()
	}
}

func generateKeys() {
	log.Println("Generating a new pair of RSA keys")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}
	pub := priv.Public()

	// priv and pub are our keys. Let's encode them

	privPEM = pem.EncodeToMemory(
		&pem.Block{Type: string(db.BetulaMetaPrivateKey), Bytes: x509.MarshalPKCS1PrivateKey(priv)},
	)
	pubPEM = pem.EncodeToMemory(
		&pem.Block{Type: string(db.BetulaMetaPublicKey), Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey))},
	)

	db.SetMetaEntry(db.BetulaMetaPrivateKey, privPEM)
	db.SetMetaEntry(db.BetulaMetaPublicKey, pubPEM)
}
