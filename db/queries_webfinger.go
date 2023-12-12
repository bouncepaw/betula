package db

import "git.sr.ht/~bouncepaw/betula/types"

// InsertWebfingerAcct inserts the entry, overwriting the previous record.
func InsertWebfingerAcct(acct types.WebfingerAcct) {
	mustExec(`
		replace into WebFingerAccts (Acct, ActorURL, Document, LastCheckedAt)
		values (?, ?, ?, ?)`, acct.Acct, acct.ActorURL, acct.Document, acct.LastCheckedAt)
}

// FetchCachedWebfingerAcct looks for an entry about the given acct (name@host). Return if found. The document is not returned, it's stored for a future or manual use.
func FetchCachedWebfingerAcct(acct string) (acctEntry types.WebfingerAcct, found bool) {
	rows := mustQuery(`select Acct, ActorURL, LastCheckedAt from WebFingerAccts where Acct = ?`, acct)
	for rows.Next() {
		mustScan(rows, &acctEntry.Acct, &acctEntry.ActorURL, &acctEntry.LastCheckedAt)
		found = true
	}
	return
}
