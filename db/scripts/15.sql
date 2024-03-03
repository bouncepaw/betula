create table Bookmarks (
    ID integer primary key autoincrement,
    URL text not null check (URL <> ''),
    Title text not null check (Title <> ''),
    Description text not null,
    Visibility integer check (Visibility = 0 or Visibility = 1 or Visibility = 2), -- private public unlisted
    CreationTime text not null default current_timestamp,
    DeletionTime text,

    RepostOf text,
    OriginalAuthorID text
);

insert into Bookmarks
    (ID, URL, Title, Description, Visibility, CreationTime, DeletionTime, RepostOf, OriginalAuthorID)
select
       ID, URL, Title, Description, Visibility, CreationTime, DeletionTime, RepostOf, null
from Posts;

drop table Posts;