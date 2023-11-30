package types

const ActivityType = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""

type ServerSoftwareKind string

const (
	SoftwareBetula  ServerSoftwareKind = "betula"
	SoftwareGeneral ServerSoftwareKind = "general"
)

type Actor struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Inbox             string `json:"inbox"`
	PreferredUsername string `json:"preferredUsername"`
	DisplayedName     string `json:"name"`
	Summary           string `json:"summary,omitempty"`
	//IconID            string             `json:"icon,omitempty"`
	ServerSoftware ServerSoftwareKind `json:"-"`
}
