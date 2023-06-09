package db

/*
This file hold keys used in the BetulaMeta table.
*/

type BetulaMetaKey string

const (
	BetulaMetaAdminUsername     BetulaMetaKey = "Admin username"
	BetulaMetaAdminPasswordHash BetulaMetaKey = "Admin password hash"
	BetulaMetaNetworkPort       BetulaMetaKey = "Network port"
	BetulaMetaSiteTitle         BetulaMetaKey = "Site title HTML"
	BetulaMetaSiteName          BetulaMetaKey = "Site name plaintext"
	BetulaMetaSiteDescription   BetulaMetaKey = "Site description Mycomarkup"
	BetulaMetaSiteURL           BetulaMetaKey = "WWW URL"
	BetulaMetaCustomCSS         BetulaMetaKey = "Custom CSS"
)
