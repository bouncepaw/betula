// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	_ "github.com/mattn/go-sqlite3"
)

/*
This file contains testing things that are used in all tests in this package, and beyond.

Pay attention to the fish below and do not forget to InitInMemoryDB!
*/

const pufferfish = "🐡"
const tropicfish = "🐠"

// InitInMemoryDB initializes a database in :memory:. Use it instead of a real db in a file for tests.
func InitInMemoryDB() {
	Initialize(":memory:")
	const q = `
insert into Bookmarks
   (URL, Title, Description, Visibility, CreationTime, DeletionTime)
values
	(
	 'https://bouncepaw.com',
	 'Bouncepaw website',
	 'A cute website by Bouncepaw',
	 0, '2023-03-17 13:13:13', null
	),
   (
    'https://mycorrhiza.wiki',
    'Mycorrhiza Wiki',
    'A wiki engine',
    1, '2023-03-17 13:14:15', null
   ),
	(
	 'http://lesarbr.es',
	 'Les Arbres',
	 'Legacy mirror of [[1]]',
	 1, '2023-03-17 20:20:20', '2023-03-18 12:45:04'
	)
`
	mustExec(q)
}

func MoreTestingBookmarks() {
	mustExec(`
insert into Bookmarks (URL, Title, Description, Visibility, CreationTime, DeletionTime)
values 
('https://1.bouncepaw', 'Uno', '', 1, '2023-03-19 12:00:00', null),
('https://2.bouncepaw', 'Dos', '', 1, '2023-03-19 14:14:14', '2023-03-19 14:14:15'),
('https://3.bouncepaw', 'Tres', '', 1, '2023-03-20 19:19:19', null),
('https://4.bouncepaw', 'Cuatro', '', 1, '2023-03-20 20:20:20', null);
`)
}
