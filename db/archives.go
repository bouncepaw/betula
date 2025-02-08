package db

import (
	"context"
	"errors"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
)

type ArtifactsRepo interface {
	Fetch(string) (*types.Artifact, error)
}

type dbArtifactsRepo struct{}

func (repo *dbArtifactsRepo) Fetch(id string) (*types.Artifact, error) {
	var artifact = types.Artifact{
		ID: id,
	}
	var row = db.QueryRow(`select MimeType, Data, IsGzipped from Artifacts where ID = ?`, id)
	var err = row.Scan(&artifact.MimeType, &artifact.Data, &artifact.IsGzipped)
	return &artifact, err
}

func NewArtifactsRepo() ArtifactsRepo {
	return &dbArtifactsRepo{}
}

type ArchivesRepo interface {
	// Store stores a new archive for the given bookmark with
	// the given artifact. It returns id of the new archive.
	Store(bookmarkID int64, artifact *types.Artifact) (int64, error)
	FetchForBookmark(bookmarkID int64) ([]types.Archive, error)
}

type dbArchivesRepo struct{}

func (repo *dbArchivesRepo) Store(bookmarkID int64, artifact *types.Artifact) (int64, error) {
	var tx, err = db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	// If the hash (ID) is taken already, it means we already have such an
	// artifact in our database. OK, whatever, that's why we `ignore`
	// in the request below. The archive will reuse the old artifact then.
	_, err = tx.Exec(`
		insert or ignore into Artifacts (ID, MimeType, Data, IsGzipped)
		values (?, ?, ?, ?)`,
		artifact.ID, artifact.MimeType, artifact.Data, artifact.IsGzipped)
	if err != nil {
		return 0, errors.Join(err, tx.Rollback())
	}

	var newArchiveID int64
	var row = tx.QueryRow(
		`insert into Archives (BookmarkID, ArtifactID) values (?, ?) returning ID`,
		bookmarkID, artifact.ID)
	err = row.Scan(&newArchiveID)
	if err != nil {
		return 0, errors.Join(err, tx.Rollback())
	}

	return newArchiveID, tx.Commit()
}

// TODO: when implementing deletion, delete artifacts iff they only have one usage

func (repo *dbArchivesRepo) FetchForBookmark(bookmarkID int64) ([]types.Archive, error) {
	var archives []types.Archive
	// Not fetching the binary data
	var rows, err = db.Query(`
		select
		    arc.ID, arc.SavedAt,
		    art.ID, art.MimeType, art.IsGzipped
		from 
		    Archives arc
		join
			Artifacts art
		on
			arc.ArtifactID = art.ID
		where
		    arc.BookmarkID = ?
		order by
		    arc.SavedAt desc
	`, bookmarkID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			archive  types.Archive
			artifact types.Artifact
		)
		err = rows.Scan(&archive.ID, &archive.SavedAt,
			&artifact.ID, &artifact.MimeType, &artifact.IsGzipped)
		if err != nil {
			return nil, err
		}

		archive.Artifact = artifact
		archives = append(archives, archive)
	}

	slog.Debug("Fetched archives for bookmark",
		"bookmarkID", bookmarkID,
		"archivesLen", len(archives))
	return archives, nil
}

func NewArchivesRepo() ArchivesRepo {
	return &dbArchivesRepo{}
}
