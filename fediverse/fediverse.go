// Package fediverse has some of the Fediverse-related functions.
package fediverse

import (
	"net/http"
	"time"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}
