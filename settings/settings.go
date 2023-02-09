// Package settings provides intermediary access to settings of a Betula instance. It handles DB interaction for you. Make sure access to it is initialized.
package settings

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
)

const (
	biggestPort = 65535
)

var cache types.Settings

// Index reads all settings from the db.
func Index() {
	networkPort := db.MetaEntry[sql.NullInt64](db.BetulaMetaNetworkPort)
	if networkPort.Valid && networkPort.Int64 > 0 && networkPort.Int64 <= biggestPort {
		cache.NetworkPort = uint(networkPort.Int64)
	} else if networkPort.Valid {
		log.Printf("An invalid network port is provided: %d. Using 1738 instead.\n", networkPort.Int64)
		cache.NetworkPort = 1738
	} else {
		cache.NetworkPort = 1738
	}

	siteTitle := db.MetaEntry[sql.NullString](db.BetulaMetaSiteTitle)
	if siteTitle.Valid {
		cache.SiteTitle = template.HTML(siteTitle.String)
	} else {
		cache.SiteTitle = "Betula"
	}
}

func NetworkPort() uint        { return cache.NetworkPort }
func SiteTitle() template.HTML { return cache.SiteTitle }

func SetSettings(settings types.Settings) {
	db.SetMetaEntry(db.BetulaMetaNetworkPort, settings.NetworkPort)
	db.SetMetaEntry(db.BetulaMetaSiteTitle, string(settings.SiteTitle))
	Index()
}
