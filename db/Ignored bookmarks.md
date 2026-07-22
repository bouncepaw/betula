# On ignoring bookmarks
Some bookmarks are not seen. Deleted bookmarks are never seen. Private bookmarks are not seen by the unauthorized. This document lists some techniques on how to provide this ignoring in the SQL queries to Betula. There is currently no consensus on which approach is the best. Judge.

## With ignored bookmarks
This is the first approach we came up with.

```sqlite
with
   IgnoredPosts as (
      -- Ignore deleted bookmarks always
      select ID from Posts where DeletionTime is not null
      union
      -- Ignore private bookmarks if so desired
      select ID from Posts where Visibility = 0 and not ?
   )
select
   CatName, 
   count(PostID)
from
   CategoriesToPosts
where
   PostID not in IgnoredPosts
group by
	CatName;
```

1. Authorization flag is passed to `?`. It is true if authorized.
2. Take deleted bookmarks.
3. Take private bookmarks if not authorized.
4. Union 2 and 3, these are ignored bookmarks.
5. Filter them out in your query.

### Another approach
We can use `JOIN`.
Note that it is more preferable to filter the bookmarks first and then join the tables.

```sqlite
select
   CatName, 
   count(PostID)
from
   CategoriesToPosts
inner join 
    (select ID from main.Posts where DeletionTime is null and (Visibility = 1 or ?)) 
as
    Filtered
on 
    CategoriesToPosts.PostID = Filtered.ID
group by
    CatName;
```

## With non-ignored bookmarks
This is the positive version of the previous approach. Not used now.

## Short condition
This one does not use the `with` expression.

```sqlite
select min(CreationTime)
from Posts
where DeletionTime is null and (Visibility = 1 or ?);
```

The `(Visibility = 1 or ?)` part needs some explanation. Consider the following table:

| Authorized? | Public? | Should be shown? |
| ----------- | ------- | ---------------- |
| 0 | 0 | 0 |
| 0 | 1 | 1 |
| 1 | 0 | 1 |
| 1 | 1 | 1 |

This table is the logical table for OR. One can also think about the logical implication and come up with a funnier way of ignoring bookmarks.