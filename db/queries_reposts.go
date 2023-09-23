package db

import "git.sr.ht/~bouncepaw/betula/types"

// RepostsFor returns all reposts known about the specified post.
func RepostsFor(id int) (reposts []types.RepostInfo) {
	const q = `
select RepostURL, ReposterName, RepostedAt from KnownReposts where PostID = ?;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var repost types.RepostInfo
		mustScan(rows, &repost.URL, &repost.Name, &repost.Timestamp)
		reposts = append(reposts, repost)
	}
	return reposts
}

func SaveRepost(postId int, repost types.RepostInfo) {
	const q = `
insert into KnownReposts (RepostURL, PostID, ReposterName)
values (?, ?, ?)`
	mustExec(repost.URL, repost.URL, postId, repost.Name)
}
