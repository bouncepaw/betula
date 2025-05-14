package db

// BetulaMetaKey is key from the BetulaMeta table.
type BetulaMetaKey string

const (
	BetulaMetaAdminUsername     BetulaMetaKey = "Admin username"
	BetulaMetaAdminPasswordHash BetulaMetaKey = "Admin password hash"
	BetulaMetaNetworkHost       BetulaMetaKey = "Network hostname"
	BetulaMetaNetworkPort       BetulaMetaKey = "Network port"
	BetulaMetaSiteTitle         BetulaMetaKey = "Site title HTML"
	BetulaMetaSiteName          BetulaMetaKey = "Site name plaintext"
	BetulaMetaSiteDescription   BetulaMetaKey = "Site description Mycomarkup"
	BetulaMetaSiteURL           BetulaMetaKey = "WWW URL"
	BetulaMetaCustomCSS         BetulaMetaKey = "Custom CSS"
	BetulaMetaPrivateKey        BetulaMetaKey = "Private key PEM"
	BetulaMetaEnableFederation  BetulaMetaKey = "Federation enabled"
	BetulaMetaPublicCustomJS    BetulaMetaKey = "Public custom JS"
	BetulaMetaPrivateCustomJS   BetulaMetaKey = "Private custom JS"
)
