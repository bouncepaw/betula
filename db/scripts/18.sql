-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

create table Notifications
(
    ID        integer primary key autoincrement,
    CreatedAt text not null default current_timestamp,
    Kind      text not null,
    Payload   blob not null -- JSON, schema varies by Kind
);