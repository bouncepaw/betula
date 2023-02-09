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
)
