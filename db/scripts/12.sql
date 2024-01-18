-- Adding primary keys from 9.sql.

-- Accounts I am following.
create table NewFollowing (
    -- ActivityPub URL ID.
    ActorID text not null primary key,
    -- When I asked to subscribe.
    SubscribedAt text not null default current_timestamp,
    -- 0 for requested, 1 for accepted. When a Reject is received, drop the entry manually.
    AcceptedStatus integer not null default 0
);

-- Accounts that follow me. Rejected accounts don't make it here.
create table NewFollowers (
    -- ActivityPub URL ID.
    ActorID text not null primary key,
    -- When the request was accepted.
    SubscribedAt text not null default current_timestamp
);

create table NewActors (
    ID text not null primary key,
    PreferredUsername text not null,
    Inbox text not null,
    DisplayedName text not null,
    Summary text not null,

    Domain text not null,
    LastCheckedAt text not null default current_timestamp
);

insert or ignore into NewFollowing select * from Following;
insert or ignore into NewFollowers select * from Followers;
insert or ignore into NewActors select * from Actors;

drop table Following;
drop table Followers;
drop table Actors;

alter table NewFollowing rename to Following;
alter table NewFollowers rename to Followers;
alter table NewActors rename to Actors;
