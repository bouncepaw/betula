package db

import (
	"context"
	"errors"
	"git.sr.ht/~bouncepaw/betula/types"
)

type ArtifactsRepo interface {
	Store(*types.Artifact) error
	Fetch(string) (*types.Artifact, error)
	Delete(string) error
}

type dbArtifactsRepo struct{}

// Boringly banal CRUD without the U.

func (repo *dbArtifactsRepo) Store(artifact *types.Artifact) error {
	var _, err = db.Exec(`insert into Artifacts (ID, MimeType, Data) values (?, ?, ?)`,
		artifact.ID, artifact.MimeType, artifact.Data)
	return err
}

func (repo *dbArtifactsRepo) Fetch(id string) (*types.Artifact, error) {
	var artifact = types.Artifact{
		ID: id,
	}
	var row = db.QueryRow(`select MimeType, Data, SavedAt, LastCheckedAt from Artifacts where ID = ?`, id)
	var err = row.Scan(&artifact.MimeType, &artifact.Data, &artifact.SavedAt, &artifact.LastCheckedAt)
	return &artifact, err
}

func (repo *dbArtifactsRepo) Delete(id string) error {
	var _, err = db.Exec(`delete from Artifacts where ID = ?`, id)
	return err
}

func NewArtifactsRepo() ArtifactsRepo {
	return &dbArtifactsRepo{}
}

type ArchivesRepo interface {
	Store(bookmarkID int64, artifact *types.Artifact) error
	AddNote(archiveID int64, note string) error
	Fetch(archiveID int64) (*types.Archive, error)
	Delete(archiveID int64) error
}

type dbArchivesRepo struct{}

func (repo *dbArchivesRepo) Store(bookmarkID int64, artifact *types.Artifact) error {
	var tx, err = db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`insert into Artifacts (ID, MimeType, Data) values (?, ?, ?)`,
		artifact.ID, artifact.MimeType, artifact.Data)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	_, err = tx.Exec(`insert into Archives (BookmarkID, ArtifactID) values (?, ?)`,
		bookmarkID, artifact.ID)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func (repo *dbArchivesRepo) AddNote(archiveID int64, note string) error {
	var _, err = db.Exec(`update Archives set Note = ? where ID = ?`,
		note, archiveID)
	return err
}

func (repo *dbArchivesRepo) Fetch(archiveID int64) (*types.Archive, error) {
	var tx, err = db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	var archive types.Archive
	var row = tx.QueryRow(`select ID, ArtifactID, Note from Archives where ID = ?`,
		archiveID)
	err = row.Scan(&archive.ID, &archive.Artifact.ID, &archive.Note)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}

	row = tx.QueryRow(`select MimeType, Data, SavedAt, LastCheckedAt from Artifacts where ID = ?`,
		archive.Artifact.ID)
	err = row.Scan(
		&archive.Artifact.MimeType, &archive.Artifact.Data,
		&archive.Artifact.SavedAt, &archive.Artifact.LastCheckedAt)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}
	return &archive, tx.Commit()
}

func (repo *dbArchivesRepo) Delete(archiveID int64) error {
	var tx, err = db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	var row = tx.QueryRow(`delete from Archives where ID = ? returning ArtifactID`,
		archiveID)
	var artifactID string
	err = row.Scan(&artifactID)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	_, err = tx.Exec(`delete from main.Artifacts where ID = ?`, artifactID)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func NewArchivesRepo() ArchivesRepo {
	return &dbArchivesRepo{}
}
