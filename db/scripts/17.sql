create table Archives (
    ID         integer primary key autoincrement,
    BookmarkID integer not null,
    ArtifactID text    not null,
    Note       text
);