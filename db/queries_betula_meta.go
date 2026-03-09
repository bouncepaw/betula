// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

func MetaEntry[T any](key BetulaMetaKey) T {
	const q = `select Value from BetulaMeta where Key = ? limit 1;`
	return querySingleValue[T](q, key)
}

func SetMetaEntry[T any](key BetulaMetaKey, val T) {
	const q = `insert or replace into BetulaMeta values (?, ?);`
	mustExec(q, key, val)
}
