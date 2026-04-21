// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import "git.sr.ht/~bouncepaw/betula/ports/settings"

func MetaEntry[T any](key settingsports.BetulaMetaKey) T {
	const q = `select Value from BetulaMeta where Key = ? limit 1;`
	return querySingleValue[T](q, key)
}

func SetMetaEntry[T any](key settingsports.BetulaMetaKey, val T) {
	const q = `insert or replace into BetulaMeta values (?, ?);`
	mustExec(q, key, val)
}
