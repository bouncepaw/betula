package types

const ActivityType = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
const OtherActivityType = "application/activity+json"

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
	PublicKey         struct {
		ID           string `json:"id"`
		Owner        string `json:"owner"`
		PublicKeyPEM string `json:"publicKeyPem"`
	} `json:"publicKey,omitempty"`
	//IconID            string             `json:"icon,omitempty"`
	ServerSoftware ServerSoftwareKind `json:"-"`

	SubscriptionStatus SubscriptionRelation `json:"-"` // Set manually
	Acct               string               `json:"-"` // Set manually
}

type SubscriptionRelation string

const (
	SubscriptionNone          SubscriptionRelation = ""
	SubscriptionTheyFollow    SubscriptionRelation = "follower"
	SubscriptionIFollow       SubscriptionRelation = "following"
	SubscriptionMutual        SubscriptionRelation = "mutual"
	SubscriptionPending       SubscriptionRelation = "pending"
	SubscriptionPendingMutual SubscriptionRelation = "pending mutual" // yours pending, theirs accepted
)

func (sr SubscriptionRelation) IsPending() bool {
	return sr == SubscriptionPending || sr == SubscriptionPendingMutual
}

func (sr SubscriptionRelation) TheyFollowUs() bool {
	return sr == SubscriptionTheyFollow || sr == SubscriptionMutual || sr == SubscriptionPendingMutual
}
