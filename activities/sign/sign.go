package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"git.sr.ht/~bouncepaw/betula/activities/httpsig"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
	"log"
	"net/http"
)

func SignRequest(rq *http.Request, content []byte) {
	keyId := settings.SiteURL() + "#main-key"
	httpsig.SignRequest(keyId, privateKey, rq, content)
}

var (
	privateKey httpsig.PrivateKey
	publicKey  httpsig.PublicKey
)

func PublicKey() string {
	s, err := httpsig.EncodeKey(publicKey.Key)
	if err != nil {
		log.Println(err)
	}
	return s // oh whatever
}

func EnsureKeys() {
	privPEM := db.MetaEntry[[]byte](db.BetulaMetaPrivateKey)
	pubPEM := db.MetaEntry[[]byte](db.BetulaMetaPublicKey)
	if privPEM == nil || pubPEM == nil {
		generateKeys() // calls for setKeys
	} else {
		setKeys(privPEM, pubPEM)
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

	privPEM := pem.EncodeToMemory(
		&pem.Block{Type: string(db.BetulaMetaPrivateKey), Bytes: x509.MarshalPKCS1PrivateKey(priv)},
	)
	pubPEM := pem.EncodeToMemory(
		&pem.Block{Type: string(db.BetulaMetaPublicKey), Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey))},
	)

	db.SetMetaEntry(db.BetulaMetaPrivateKey, privPEM)
	db.SetMetaEntry(db.BetulaMetaPublicKey, pubPEM)
	setKeys(privPEM, pubPEM)
}

func setKeys(privPEM, pubPEM []byte) {
	privateKey = httpsig.PrivateKey{
		Type: httpsig.RSA,
		Key:  privPEM,
	}
	publicKey = httpsig.PublicKey{
		Type: httpsig.RSA,
		Key:  pubPEM,
	}
}
