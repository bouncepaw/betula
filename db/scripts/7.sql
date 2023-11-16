-- Artifacts is a storage for binary artifacts from remote resources
-- such as avatars, favicons, whatnot. General purpose.
create table Artifacts
(
    ID            text primary key,
    MimeType      text, -- nullable. Just send as is if no idea what is it.
    Data          blob,
    SavedAt       text not null default current_timestamp,
    LastCheckedAt text -- null = never checked
);

-- Actors is a storage for all known actors.
create table Actors
(
    ID                text primary key, -- ActivityPub's URL id
    Inbox             text not null,    -- The spec says it MUST be present, so we'll find it somehow, don't worry.
    PreferredUsername text not null,
    DisplayedName     text not null,
    Summary           text not null,
    IconID            text,             -- ActivityPub's URL id, nullable
    ServerSoftware    text,             -- "betula", "mastodon", whatever.
    foreign key (IconID) references Artifacts (ID)
);

-- Subscriptions lists all known subscriptionship relations between you and others.
create table Subscriptions
(
    ID           integer primary key, -- Will be used for referencing subscription filters
    AuthorID     text,                -- ActivityPub's URL id. If null, then it means your Betula
    SubscriberID text,                -- ActivityPub's URL id. If null, then it means your Betula
    SubscribedAt text    not null default current_timestamp,
    Accepted     integer not null,    -- 0 = not accepted, 1 = accepted
    foreign key (AuthorID) references Actors (ID),
    foreign key (SubscriberID) references Actors (ID)
);

-- IncomingPosts lists all known posts that were sent from who we follow our way.
-- These posts can and will be deleted at user's will.
--
-- We don't have Visibility and DeletedAt fields. If we have the post, it's public
-- enough for our purposes. DeletedAt bears no value, we'll just drop deleted
-- posts.
create table IncomingPosts
(
    ID          text primary key, -- ActivityPub's URL, serves as post's URL.
    RepostOf    text,             -- nullable. Distinguishes if this is an original post or not

    Title       text not null,
    Description text not null,
    CreatedAt   text not null
);
