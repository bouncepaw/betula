package httpsig

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

var pubkey = PublicKey{}
var privkey = PrivateKey{}

func TestSignPost(t *testing.T) {
	msg := []byte("hello")
	req, _ := http.NewRequest("POST", "https://example.com/url", bytes.NewReader(msg))
	req.Header.Set("Content-Type", "text/plain")

	SignRequest("nameofkey", privkey, req, msg)

	fmt.Printf("POST Signature: %s\n", req.Header.Get("Signature"))

	_, err := VerifyRequest(req, msg, func(s string) (PublicKey, error) {
		return pubkey, nil
	})
	if err != nil {
		t.Errorf("verify: %s", err)
	}
	_, err = VerifyRequest(req, []byte("false"), func(s string) (PublicKey, error) {
		return pubkey, nil
	})
	if err == nil {
		t.Error("bad verify")
	}
}

func TestSignGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/url", nil)

	SignRequest("nameofkey", privkey, req, nil)

	fmt.Printf("GET Signature: %s\n", req.Header.Get("Signature"))

	_, err := VerifyRequest(req, nil, func(s string) (PublicKey, error) {
		return pubkey, nil
	})
	if err != nil {
		t.Errorf("verify: %s", err)
	}
}

func init() {
	var err error
	privkey, pubkey, err = DecodeKey(privkeyString)
	if err != nil {
		panic(err)
	}
}

var pubkeyString = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt0c14FLZWHqgNJzlHrpR
yx9CuW+w5DYo/5s59FZHEgNzTjUJyE3Jfxx1j/6gcCWmZGoTULR5ILZySe/jagQw
JvLl7dkOcf7K3FGh3JpoTeZpInFjVSyKag3PfstQI/Hq/JIV9mysGk2hNoo/0Jvh
2vT48jGuxhrwHchJdNeYs8eBtvSfIlXsnt9K1qZVU9T+anG7wwn42t5GcQ391vxk
DLgUzF5K1wSzOQoXh5lryW/BZzEmAhFqGZz69UsjRoS1ia53nb1LjXBpZdetvqU6
0YlOHfRc19tgUnmkMqXqMV6lRUmFqyFh/kj50GE5FsuQSjlNaKiotuZg/KJ/DhfO
HwIDAQAB
-----END PUBLIC KEY-----
`
var privkeyString = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAt0c14FLZWHqgNJzlHrpRyx9CuW+w5DYo/5s59FZHEgNzTjUJ
yE3Jfxx1j/6gcCWmZGoTULR5ILZySe/jagQwJvLl7dkOcf7K3FGh3JpoTeZpInFj
VSyKag3PfstQI/Hq/JIV9mysGk2hNoo/0Jvh2vT48jGuxhrwHchJdNeYs8eBtvSf
IlXsnt9K1qZVU9T+anG7wwn42t5GcQ391vxkDLgUzF5K1wSzOQoXh5lryW/BZzEm
AhFqGZz69UsjRoS1ia53nb1LjXBpZdetvqU60YlOHfRc19tgUnmkMqXqMV6lRUmF
qyFh/kj50GE5FsuQSjlNaKiotuZg/KJ/DhfOHwIDAQABAoIBAQCoilCuICIE7odi
uqEsV7Sd8PpgIqjtxCyBmdJ6sdibZRbk5XtAeuAAB0DJESOi3cyc7LskbUIyZfTF
r1dXB9DsEFSHHCLfi6orXtpVTpz6fhdSeCkbi2Eh099rPzZMR8yLRR/zQ84kRh0p
VYsHoEHbI4nG1w9c2CrViHicfSLMtuvz2J2rvyYjNSU7Ar0PdbC3uNLaxKqA2kgr
qmtSuptjBauN/Sup85/KnKUERuKP8LRs0Psqh2QKyN6ZdIbU/CZEfbp3W4kuyxc7
YQHU40s3Pj93Wn1zXaPNioWOFAswNJVMdPp7Nozc0biS9bBUxwiFifChRpPTOLaU
/u5juFFhAoGBAMcmCo5mBYbpor7OCUlmUJheKGZSq6m3XjWjgQcMUFkKYloFaODb
FuH0HDN9Y1neBpy1WSE3UQmS+w8zw1Jt8eW7OqYdgzjLARaUf4vIHFUvFKuw6uL+
BMOfL3J/NAE+eNw1/Hz86Z815JCexYZL6xypmaqAuLT/m98nkxoaYBFxAoGBAOuZ
V/22az5gW0K/DhimJbYBk1cQNaTmGYhzLmM6XPDOZbCz83loPuiKsD2E9ZKX9UAC
hi1X3kqoXUeYcyqMUD50KF++0meAQf3DRTmYAjvkQ6VwdOcbuIQXn/gFs0tuoszd
l6eWrTPmMBS+zpgIqccFd8DdVGtxvo6RfAmrRxCPAoGAWmuPR3BS+hqCZheuZ8Eo
vsWhmjPW9UvoXnpKTyTsJkFsvmrOX6maDiWD2G0J+vewEN7WBRrUlIBDtXdPK9H3
jtMfoeSse9DQQaxS7OiC1Lp3rCy7uSyUhS11oYrX1ejDf1iTtzwt5rfVe0Rbcspt
iaoHtz6SnrufzgZt5+Ap1kECgYAsxG2Q2ynTp3GP5EfkbSW7SN9baswWslZltCU7
W6qvYzi1c+wuxJ03iKrmda5IFbHXYONoGEs3+ngHE7PGgPT6eQ3264aFfjyL4J/1
yqmaAczM0eqUw5KzHt4ZvdOM4M/0h6K6iIoO042NU5hkETlZhPN1ZVkWNX3VD1X3
bGFLhwKBgQCFYlJaiX1a7/34vSPyBi+DX3TtSWSoKoC4WnqxLWjGsCiOmMbITNdr
sGaeQDAMwHLuuXiiT+FX1OXpQSP1PGdM2tbaCPTiwB0JnE0Azw3820jkERijkqzH
dBxkc3/xi4/CHznq8GEyoKuevcCFh2o55+H8LolTf0/ImhJzlRI9kQ==
-----END RSA PRIVATE KEY-----
`
