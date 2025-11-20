-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

drop table Subscriptions;

-- DROPPED IN 12: they were migrated to similar tables that have primary keys

-- Accounts I am following.
create table Following (
    -- ActivityPub URL ID.
    ActorID text not null,
    -- When I asked to subscribe.
    SubscribedAt text not null default current_timestamp,
    -- 0 for requested, 1 for accepted. When a Reject is received, drop the entry manually.
    AcceptedStatus integer not null default 0
);

-- Accounts that follow me. Rejected accounts don't make it here.
create table Followers (
    -- ActivityPub URL ID.
    ActorID text not null,
    -- When the request was accepted.
    SubscribedAt text not null default current_timestamp
);
