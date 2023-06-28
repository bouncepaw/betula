// Package settings provides intermediary access to settings of a Betula instance. It handles DB interaction for you. Make sure access to it is initialized.
package settings

import (
	"database/sql"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/types"
	"html"
	"html/template"
	"log"
	"net/url"
)

type Uintport uint
type NullInt64Port sql.NullInt64

const biggestPort = 65535
const defaultBetulaPort = 1738

var cache types.Settings

// Those that did not fit in cache go in their own variables below. Handle with thought.
var cacheSiteDescription template.HTML

func (port Uintport) ValidatePort() uint {
	if port > 0 && port <= biggestPort {
		return uint(port)
	} else {
		log.Printf("An invalid network port is provided: %d. Using %d instead.\n", port, defaultBetulaPort)
		return defaultBetulaPort
	}
}

func (port NullInt64Port) ValidatePort() uint {
	if port.Valid && port.Int64 > 0 && port.Int64 <= biggestPort {
		return uint(port.Int64)
	} else if port.Valid && db.PostCount(true) > 0 {
		log.Printf("An invalid network port is provided: %d. Using %d instead.\n", port.Int64, defaultBetulaPort)
		return defaultBetulaPort
	} else {
		return defaultBetulaPort
	}
}

// Index reads all settings from the db.
func Index() {
	networkPort := db.MetaEntry[sql.NullInt64](db.BetulaMetaNetworkPort)
	cache.NetworkPort = NullInt64Port(networkPort).ValidatePort()

	siteName := db.MetaEntry[sql.NullString](db.BetulaMetaSiteName)
	if siteName.Valid && siteName.String != "" {
		cache.SiteName = siteName.String
	} else {
		cache.SiteName = "Betula"
	}

	siteTitle := db.MetaEntry[sql.NullString](db.BetulaMetaSiteTitle)
	if siteTitle.Valid && siteTitle.String != "" {
		cache.SiteTitle = template.HTML(siteTitle.String)
	} else {
		cache.SiteTitle = template.HTML(html.EscapeString(cache.SiteName))
	}

	siteDescription := db.MetaEntry[sql.NullString](db.BetulaMetaSiteDescription)
	if siteDescription.Valid && siteDescription.String != "" {
		cache.SiteDescriptionMycomarkup = siteDescription.String
		cacheSiteDescription = myco.MarkupToHTML(siteDescription.String)
	} else {
		cache.SiteDescriptionMycomarkup = ""
		cacheSiteDescription = ""
	}

	siteURL := db.MetaEntry[sql.NullString](db.BetulaMetaSiteURL)
	if !siteURL.Valid {
		cache.SiteURL = fmt.Sprintf("http://localhost:%d", cache.NetworkPort)
	} else {
		addr, err := url.ParseRequestURI(siteURL.String)
		if err != nil || addr.Path != "" {
			cache.SiteURL = fmt.Sprintf("http://localhost:%d", cache.NetworkPort)
		} else {
			cache.SiteURL = siteURL.String
		}
	}
}

func SiteURL() string                    { return cache.SiteURL }
func NetworkPort() uint                  { return cache.NetworkPort }
func SiteName() string                   { return cache.SiteName }
func SiteTitle() template.HTML           { return cache.SiteTitle }
func SiteDescriptionHTML() template.HTML { return cacheSiteDescription }
func SiteDescriptionMycomarkup() string  { return cache.SiteDescriptionMycomarkup }

func SetSettings(settings types.Settings) {
	if settings.SiteName == "" {
		settings.SiteName = "Betula"
	}
	db.SetMetaEntry(db.BetulaMetaNetworkPort, settings.NetworkPort)
	db.SetMetaEntry(db.BetulaMetaSiteName, settings.SiteName)
	db.SetMetaEntry(db.BetulaMetaSiteTitle, string(settings.SiteTitle))
	db.SetMetaEntry(db.BetulaMetaSiteDescription, settings.SiteDescriptionMycomarkup)
	db.SetMetaEntry(db.BetulaMetaSiteURL, settings.SiteURL)
	Index()
}

func (port Uintport) SetNetworkPort() {
	db.SetMetaEntry(db.BetulaMetaNetworkPort, port.ValidatePort())
}

func (port NullInt64Port) SetNetworkPort() {
	db.SetMetaEntry(db.BetulaMetaNetworkPort, port.ValidatePort())
}
