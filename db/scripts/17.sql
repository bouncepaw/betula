drop table if exists Artifacts;
drop table if exists Archives;

-- Artifacts are icons, compressed web pages, etc.
create table Artifacts
(
    ID        text primary key,
    MimeType  text not null,
    Data      blob not null,
    IsGzipped integer not null default 0
);

-- Archives are copies of web pages.
create table Archives
(
    ID         integer primary key autoincrement,
    BookmarkID integer not null,
    ArtifactID text    not null,
    SavedAt    text    not null default current_timestamp
);
