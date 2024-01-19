-- Sorry for losing data here, my dear adventurous beta user:
drop table if exists Actors;
drop table if exists Following;
drop table if exists Followers;

-- Adding primary keys to the tables from 9.sql.

-- Accounts I am following.
create table Following (
    -- ActivityPub URL ID.
    ActorID text not null primary key,
    -- When I asked to subscribe.
    SubscribedAt text not null default current_timestamp,
    -- 0 for requested, 1 for accepted. When a Reject is received, drop the entry manually.
    AcceptedStatus integer not null default 0
);

-- Accounts that follow me. Rejected accounts don't make it here.
create table Followers (
    -- ActivityPub URL ID.
    ActorID text not null primary key,
    -- When the request was accepted.
    SubscribedAt text not null default current_timestamp
);

-- See 7 for the prev ver.
create table Actors (
    ID text not null primary key,
    PreferredUsername text not null,
    Inbox text not null,
    DisplayedName text not null,
    Summary text not null,

    Domain text not null,
    LastCheckedAt text not null default current_timestamp
);

