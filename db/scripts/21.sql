-- SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- Fix dangling view from v0.4 era.
drop view if exists Categories;

alter table Bookmarks rename column RepostOf to RemarkedID;
alter table Bookmarks add column RemarkText text;
