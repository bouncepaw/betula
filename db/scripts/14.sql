-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- DROPPED IN 20: Remark, web url and non-Mycomarkup description support
-- See 13.
-- The default value is not to be used, I just have to provide some default value for not null column.
alter table RemoteBookmarks add column URL text not null default '';
-- END DROPPEd