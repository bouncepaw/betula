-- Artifacts became immutable and they could not expire anymore.
alter table Artifacts drop column LastCheckedAt;

-- Used in archives.
alter table Artifacts add column IsGzipped integer not null default 0;

-- URLs in bookmarks can be changed, so we should record them too.
alter table Archives add column URL text not null default '';