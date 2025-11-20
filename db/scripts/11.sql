-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

drop table Actors;
drop table WebFingerAccts;

-- DROPPED IN 12: forgot the primary key

create table Actors (
    ID text not null,
    PreferredUsername text not null,
    Inbox text not null,
    DisplayedName text not null,
    Summary text not null,

    Domain text not null,
    LastCheckedAt text not null default current_timestamp
);
