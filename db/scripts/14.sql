-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- See 13.
-- The default value is not to be used, I just have to provide some default value for not null column.
alter table RemoteBookmarks add column URL text not null default '';