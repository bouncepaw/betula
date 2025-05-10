package db

import "strconv"

type GoLinksRepo interface {
	GetBookmarkID(slug string, authorized bool) (int64, error)
	Assign(bookmarkID int64, slug string) error
}

type goLinksRepo struct{}

func (g *goLinksRepo) GetBookmarkID(slug string, authorized bool) (int64, error) {
	var bid, err = strconv.ParseInt(slug, 10, 64)
	// If is a bookmark id:
	if err == nil {
		var res = db.QueryRow(
			`select Visibility from Bookmarks where ID = ?`, bid)
		var visibility int64
		err = res.Scan(&visibility)
		if err != nil {
			return 0, err
		}
		// If unauthorized and the bookmark is private:
		if !authorized && visibility == 0 {
			return 0, nil
		}
		return bid, nil
	}

	panic("attack")
}

func (g *goLinksRepo) Assign(bookmarkID int64, slug string) error {
	var _, err = db.Exec(
		`insert into GoLinks (ID, BookmarkID) values (?, ?);`,
		slug, bookmarkID)
	return err
}
