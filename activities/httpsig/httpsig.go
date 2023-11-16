// Copyright (c) 2019 Ted Unangst <tedu@tedunangst.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// An implementation of HTTP Signatures
package httpsig

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ed25519"
)

type KeyType int

const (
	None KeyType = iota
	RSA
	Ed25519
)

type PublicKey struct {
	Type KeyType
	Key  interface{}
}

func (pubkey PublicKey) Verify(msg []byte, sig []byte) error {
	switch pubkey.Type {
	case RSA:
		return rsa.VerifyPKCS1v15(pubkey.Key.(*rsa.PublicKey), crypto.SHA256, msg, sig)
	case Ed25519:
		ok := ed25519.Verify(pubkey.Key.(ed25519.PublicKey), msg, sig)
		if !ok {
			return fmt.Errorf("verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unknown key type")
	}
}

type PrivateKey struct {
	Type KeyType
	Key  interface{}
}

func (privkey PrivateKey) Sign(msg []byte) []byte {
	switch privkey.Type {
	case RSA:
		sig, err := rsa.SignPKCS1v15(rand.Reader, privkey.Key.(*rsa.PrivateKey), crypto.SHA256, msg)
		if err != nil {
			panic(fmt.Errorf("error signing msg: %s", err))
		}
		return sig
	case Ed25519:
		return ed25519.Sign(privkey.Key.(ed25519.PrivateKey), msg)
	default:
		panic("unknown key type")
	}
}

func sb64(data []byte) string {
	var sb strings.Builder
	b64 := base64.NewEncoder(base64.StdEncoding, &sb)
	b64.Write(data)
	b64.Close()
	return sb.String()
}

func b64s(s string) []byte {
	var buf bytes.Buffer
	b64 := base64.NewDecoder(base64.StdEncoding, strings.NewReader(s))
	io.Copy(&buf, b64)
	return buf.Bytes()
}

func sb64sha256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return sb64(h.Sum(nil))
}

// Sign a request and add Signature header
func SignRequest(keyname string, key PrivateKey, req *http.Request, content []byte) {
	headers := []string{"(request-target)", "date", "host"}
	if strings.ToLower(req.Method) != "get" {
		headers = append(headers, "content-type", "digest")
	}
	var stuff []string
	for _, h := range headers {
		var s string
		switch h {
		case "(request-target)":
			s = strings.ToLower(req.Method) + " " + req.URL.RequestURI()
		case "date":
			s = req.Header.Get(h)
			if s == "" {
				s = time.Now().UTC().Format(http.TimeFormat)
				req.Header.Set(h, s)
			}
		case "host":
			s = req.Header.Get(h)
			if s == "" {
				s = req.URL.Hostname()
				req.Header.Set(h, s)
			}
		case "content-type":
			s = req.Header.Get(h)
		case "digest":
			s = req.Header.Get(h)
			if s == "" {
				s = "SHA-256=" + sb64sha256(content)
				req.Header.Set(h, s)
			}
		}
		stuff = append(stuff, h+": "+s)
	}

	h := sha256.New()
	h.Write([]byte(strings.Join(stuff, "\n")))
	sig := key.Sign(h.Sum(nil))
	bsig := sb64(sig)

	sighdr := fmt.Sprintf(`keyId="%s",algorithm="%s",headers="%s",signature="%s"`,
		keyname, "rsa-sha256", strings.Join(headers, " "), bsig)
	req.Header.Set("Signature", sighdr)
}

// Verify the Signature header for a request is valid.
// The request body should be provided separately.
// The lookupPubkey function takes a keyname and returns a public key.
// Returns keyname if known, and/or error.
func VerifyRequest(req *http.Request, content []byte, lookupPubkey func(string) (PublicKey, error)) (string, error) {
	sighdr := req.Header.Get("Signature")
	if sighdr == "" {
		return "", fmt.Errorf("no signature header")
	}

	var keyname, algo, heads, bsig string
	for _, v := range strings.Split(sighdr, ",") {
		name, val, ok := strings.Cut(v, "=")
		if !ok {
			return "", fmt.Errorf("bad scan: %s from %s", v, sighdr)
		}
		val = strings.TrimPrefix(val, `"`)
		val = strings.TrimSuffix(val, `"`)
		switch name {
		case "keyId":
			keyname = val
		case "algorithm":
			algo = val
		case "headers":
			heads = val
		case "signature":
			bsig = val
		default:
			return "", fmt.Errorf("bad sig val: %s from %s", name, sighdr)
		}
	}
	if keyname == "" || algo == "" || heads == "" || bsig == "" {
		return "", fmt.Errorf("missing a sig value")
	}

	key, err := lookupPubkey(keyname)
	if err != nil {
		return keyname, err
	}
	if key.Type == None {
		return keyname, fmt.Errorf("no key for %s", keyname)
	}
	required := make(map[string]bool)
	required["(request-target)"] = true
	required["date"] = true
	required["host"] = true
	if strings.ToLower(req.Method) != "get" {
		required["digest"] = true
	}
	headers := strings.Split(heads, " ")
	var stuff []string
	for _, h := range headers {
		var s string
		switch h {
		case "(request-target)":
			s = strings.ToLower(req.Method) + " " + req.URL.RequestURI()
		case "host":
			s = req.Host
			if s == "" {
				return "", fmt.Errorf("httpsig: no host header value")
			}
		case "digest":
			s = req.Header.Get(h)
			expv := "SHA-256=" + sb64sha256(content)
			if s != expv {
				return "", fmt.Errorf("digest header '%s' did not match content", s)
			}
		case "date":
			s = req.Header.Get(h)
			d, err := time.Parse(http.TimeFormat, s)
			if err != nil {
				return "", fmt.Errorf("error parsing date header: %s", err)
			}
			now := time.Now()
			if d.Before(now.Add(-30*time.Minute)) || d.After(now.Add(30*time.Minute)) {
				return "", fmt.Errorf("date header '%s' out of range", s)
			}
		default:
			s = req.Header.Get(h)
		}
		delete(required, h)
		stuff = append(stuff, h+": "+s)
	}
	if len(required) > 0 {
		var missing []string
		for h := range required {
			missing = append(missing, h)
		}
		return "", fmt.Errorf("required httpsig headers missing (%s)", strings.Join(missing, ","))
	}

	h := sha256.New()
	h.Write([]byte(strings.Join(stuff, "\n")))
	sig := b64s(bsig)
	err = key.Verify(h.Sum(nil), sig)
	if err != nil {
		return keyname, err
	}
	return keyname, nil
}

// Unmarshall an ASCII string into (optional) private and public keys
func DecodeKey(s string) (pri PrivateKey, pub PublicKey, err error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		err = fmt.Errorf("no pem data")
		return
	}
	switch block.Type {
	case "PUBLIC KEY":
		var k interface{}
		k, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err == nil {
			pub.Key = k
			switch k.(type) {
			case *rsa.PublicKey:
				pub.Type = RSA
			case ed25519.PublicKey:
				pub.Type = Ed25519
			}
		}
	case "PRIVATE KEY":
		var k interface{}
		k, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			pri.Key = k
			switch k.(type) {
			case *rsa.PrivateKey:
				pri.Type = RSA
			case ed25519.PrivateKey:
				pri.Type = Ed25519
			}
		}
	case "RSA PUBLIC KEY":
		pub.Key, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err == nil {
			pub.Type = RSA
		}
	case "RSA PRIVATE KEY":
		var rsakey *rsa.PrivateKey
		rsakey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			pri.Key = rsakey
			pri.Type = RSA
			pub.Key = &rsakey.PublicKey
			pub.Type = RSA
		}
	default:
		err = fmt.Errorf("unknown key type")
	}
	return
}

// Marshall an RSA key into an ASCII string
func EncodeKey(i interface{}) (string, error) {
	var b pem.Block
	var err error
	switch k := i.(type) {
	case *rsa.PrivateKey:
		b.Type = "RSA PRIVATE KEY"
		b.Bytes = x509.MarshalPKCS1PrivateKey(k)
	case *rsa.PublicKey:
		b.Type = "PUBLIC KEY"
		b.Bytes, err = x509.MarshalPKIXPublicKey(k)
	case ed25519.PrivateKey:
		b.Type = "PRIVATE KEY"
		b.Bytes, err = x509.MarshalPKCS8PrivateKey(k)
	case ed25519.PublicKey:
		b.Type = "PUBLIC KEY"
		b.Bytes, err = x509.MarshalPKIXPublicKey(k)
	default:
		err = fmt.Errorf("unknown key type: %s", k)
	}
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&b)), nil
}
