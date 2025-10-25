package archivingports

import "git.sr.ht/~bouncepaw/betula/types"

type Service interface {
	Archive(types.Bookmark) (int64, error)
}

// Fetcher fetches documents.
type Fetcher interface {
	// Fetch fetches an archive copy for the document identified by URL.
	// Returns contents, MIME-type and a possible error.
	Fetch(url string) ([]byte, string, error)
}

type ArtifactsRepo interface {
	Fetch(string) (*types.Artifact, error)
}

type ArchivesRepo interface {
	// Store stores a new archive for the given bookmark with
	// the given artifact. It returns id of the new archive.
	Store(bookmarkID int64, artifact *types.Artifact) (int64, error)
	FetchForBookmark(bookmarkID int64) ([]types.Archive, error)
	DeleteArchive(archiveID int64) error
}
