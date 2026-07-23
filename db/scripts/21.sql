-- SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
--
-- SPDX-License-Identifier: AGPL-3.0-only

alter table Bookmarks rename column RepostOf to RemarkedID;
alter table Bookmarks add column RemarkText text;
