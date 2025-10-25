package db

import (
	"context"
	"errors"
	"git.sr.ht/~bouncepaw/betula/ports/archiving"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
)

type dbArtifactsRepo struct{}

func (repo *dbArtifactsRepo) Fetch(id string) (*types.Artifact, error) {
	var artifact = types.Artifact{
		ID: id,
	}
	var row = db.QueryRow(`select MimeType, Data, IsGzipped, length(Data) from Artifacts where ID = ?`, id)
	var err = row.Scan(&artifact.MimeType, &artifact.Data, &artifact.IsGzipped, &artifact.Size)
	return &artifact, err
}

func NewArtifactsRepo() archivingports.ArtifactsRepo {
	return &dbArtifactsRepo{}
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

func (repo *dbArchivesRepo) DeleteArchive(archiveID int64) error {
	var tx, err = db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Artifacts might be reused, so after deleting the archive,
	// the corresponding artifact is to be deleted only if
	// no other archives refer it.

	var artifactID string
	var row = tx.QueryRow(
		`delete from Archives where ID = ? returning ArtifactID`,
		archiveID)
	if err = row.Scan(&artifactID); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	var artifactUsageCount int64
	row = tx.QueryRow(
		`select count(*) from Archives where ArtifactID = ?`,
		artifactID)
	if err = row.Scan(&artifactUsageCount); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	if artifactUsageCount == 0 {
		_, err = tx.Exec(`delete from Artifacts where ID = ?`, artifactID)
		if err != nil {
			return errors.Join(err, tx.Rollback())
		}
	}

	return tx.Commit()
}

func (repo *dbArchivesRepo) FetchForBookmark(bookmarkID int64) ([]types.Archive, error) {
	var archives []types.Archive
	// Not fetching the binary data
	var rows, err = db.Query(`
		select
		    arc.ID, arc.SavedAt,
		    art.ID, art.MimeType, art.IsGzipped, length(art.Data)
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
			&artifact.ID, &artifact.MimeType, &artifact.IsGzipped, &artifact.Size)
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

func NewArchivesRepo() archivingports.ArchivesRepo {
	return &dbArchivesRepo{}
}
