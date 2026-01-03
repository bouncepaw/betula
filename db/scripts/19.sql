-- SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- Likes represents Like activities,
-- both incoming over ActivityPub and outgoing.
create table Likes
(
    -- ID is activity's "id" field (a URL).
    -- Null for outgoing likes.
    -- Not a primary key, but is unique.
    ID         text null unique,
    -- ActorID is activity's "actor" field (a URL).
    -- Null for outgoing likes.
    ActorID    text null,
    -- ObjectID is activity's "object" field (a URL)
    -- or a string representation of local bookmark id.
    -- We could have an untyped column,
    -- but Go side is not comfortable with that.
    ObjectID   text not null,

    -- SavedAt is when this db entry was created.
    -- Used for sorting likes.
    SavedAt    text not null default current_timestamp,

    SourceJSON text null
);

-- LikeCollections represents known remote "likes" collections data,
-- without "items" represented in db (we don't really need that).
create table LikeCollections
(
    -- ID is "id" of the collection.
    -- Null for anonymous embedded collections.
    -- Not a primary key, but is unique.
    ID            text null unique,
    LikedObjectID text primary key,
    TotalItems    integer not null,

    SourceJSON    text    null -- not stripping "items"
);
