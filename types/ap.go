package types

type ServerSoftwareKind string

const (
	SoftwareBetula  ServerSoftwareKind = "betula"
	SoftwareGeneral ServerSoftwareKind = "general"
)

type Actor struct {
	ID                string             `json:"id"`
	Inbox             string             `json:"inbox"`
	PreferredUsername string             `json:"preferredUsername"`
	DisplayedName     string             `json:"name"`
	Summary           string             `json:"summary,omitempty"`
	IconID            string             `json:"icon,omitempty"`
	ServerSoftware    ServerSoftwareKind `json:"-"`
}
