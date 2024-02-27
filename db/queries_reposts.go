package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"time"
)

// RepostsOf returns all reposts known about the specified bookmark.
func RepostsOf(id int) (reposts []types.RepostInfo, err error) {
	rows := mustQuery(`select RepostURL, ReposterName, RepostedAt from KnownReposts where PostID = ?`, id)
	for rows.Next() {
		var repost types.RepostInfo
		var timestamp string
		mustScan(rows, &repost.URL, &repost.Name, &timestamp)
		repost.Timestamp, err = time.Parse(types.TimeLayout, timestamp)
		if err != nil {
			log.Printf("When reading tags for bookmark no. %d: %s\n", id, err)
		}
		reposts = append(reposts, repost)
	}
	return reposts, nil
}

func CountRepostsOf(id int) int {
	const q = `select count(*) from KnownReposts where PostID = ?;`
	return querySingleValue[int](q, id)
}

func SaveRepost(bookmarkID int, repost types.RepostInfo) {
	const q = `
insert into KnownReposts (RepostURL, PostID, ReposterName)
values (?, ?, ?)
on conflict do nothing`
	mustExec(q, repost.URL, bookmarkID, repost.Name)
}

func DeleteRepost(bookmarkID int, repostURL string) {
	mustExec(`delete from KnownReposts where RepostURL = ? and PostID = ?`, repostURL, bookmarkID)
}
