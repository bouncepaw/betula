drop table IncomingPosts;

-- RemoteBookmarks lists all known bookmarks that were sent from who we follow our way.
-- These bookmarks can and will be deleted at some point.
--
-- We don't have Visibility and DeletedAt fields. If we have the post, it's public
-- enough for our purposes. DeletedAt bears no value, we'll just drop deleted
-- posts.
create table RemoteBookmarks (
    -- EXTENDED IN 14: field URL text not null was added.
    -- ActivityPub's ID, serves as post's URL.
    ID                    text primary key,
    -- Nullable. Distinguishes if this is an original post or not. If not null, it is ID of the original post.
    RepostOf              text,
    ActorID               text not null,

    -- If the original post didn't have a title, then it probably wasn't made with Betula in mind. Maybe come up with something yourself. Anyway, it is not null and must be present.
    Title                 text not null check (Title <> ''),
    DescriptionHTML       text not null,
    -- Nullable. Only Betula broadcasts it out after all.
    DescriptionMycomarkup text,
    -- No default value. Must be present in the activity.
    PublishedAt           text not null,
    -- Null means it was never Updated.
    UpdatedAt             text,
    Activity              blob
);

create table RemoteTags (
    Name text not null,
    BookmarkID text not null,
    unique(Name, BookmarkID)
);