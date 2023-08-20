// Package stricks (string tricks) provides common string operations that looked like they belong here.
package stricks

import "net/url"

func ValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
