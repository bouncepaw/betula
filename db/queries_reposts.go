package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"time"
)

// RepostsOf returns all reposts known about the specified post.
func RepostsOf(id int) (reposts []types.RepostInfo, err error) {
	const q = `
select RepostURL, ReposterName, RepostedAt from KnownReposts where PostID = ?;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var repost types.RepostInfo
		var timestamp string
		mustScan(rows, &repost.URL, &repost.Name, &timestamp)
		repost.Timestamp, err = time.Parse(types.TimeLayout, timestamp)
		if err != nil {
			log.Println(err)
		}
		reposts = append(reposts, repost)
	}
	return reposts, nil
}

func CountRepostsOf(id int) int {
	const q = `select count(*) from KnownReposts where PostID = ?;`
	return querySingleValue[int](q, id)
}

func SaveRepost(postId int, repost types.RepostInfo) {
	const q = `
insert into KnownReposts (RepostURL, PostID, ReposterName)
values (?, ?, ?)
on conflict do nothing`
	mustExec(q, repost.URL, postId, repost.Name)
}

func DeleteRepost(postId int, repostURL string) {
	mustExec(`delete from KnownReposts where RepostURL = ? and PostID = ?`, repostURL, postId)
}
