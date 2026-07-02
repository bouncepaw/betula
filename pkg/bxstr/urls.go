// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package bxstr

import "net/url"

func IsValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

func MustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func SameHost(s1, s2 string) bool {
	u1, err1 := url.ParseRequestURI(s1)
	u2, err2 := url.ParseRequestURI(s2)
	return err1 == nil && err2 == nil && u1.Host == u2.Host
}

func ValidURLWithQuery(s string, m map[string]string) string {
	u := MustParseURL(s)
	q := u.Query()
	for k, v := range m {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
