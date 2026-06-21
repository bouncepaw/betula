-- SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- See 13 and 14 for prev version.

create table Timeline (
	--- Common AP object attributes.
	-- ActivityPub object ID (a URL)
	ID            text primary key not null,
	-- ActorID is a foreign key.
	ActorID       text             not null,
	PublishedAt   text             not null,
	UpdatedAt     text             null,
	-- WebURL is the URL of web representation of the bookmark.
	-- For Announce remarks, it's null.
	-- For others, there's probably a web representation.
	-- If not set, fallback to ID.
	WebURL        text             null,
	-- Activity is likely a Create{Note} or Update{Note} for a bookmark,
	-- Announce, Create{Note}, Update{Note} for a remark.
	Activity      blob             null,

	--- The text part.
	HTML          text             null,
	-- SourceType is the type of source text.
	-- null for Mycomarkup, 'P' for Plain text.
	SourceType    text             null,
	-- Source is the original text for the remark/bookmark.
	-- If not set, fallback to HTML.
	-- If both Source and HTML are null, no text.
	Source        text             null,

	--- Bookmark attributes.
	-- BookmarkedURL is null for remarks.
	BookmarkedURL text             null,
	-- BookmarkTitle is null for remarks.
	BookmarkTitle text             null,

	--- Remark attributes.
	-- RemarkedID is null for bookmarks.
	RemarkedID    text             null -- AP ID
);

insert
into Timeline
(
	ID, ActorID, PublishedAt, UpdatedAt, WebURL, Activity,
	HTML, SourceType, Source,
	BookmarkedURL, BookmarkTitle,
	RemarkedID
)
select ID,
	ActorID,
	PublishedAt,
	UpdatedAt,
	null,
	Activity,
	DescriptionHTML,
	null,
	DescriptionMycomarkup,
	URL,
	Title,
	RepostOf
from RemoteBookmarks;

drop table RemoteBookmarks;
