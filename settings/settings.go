// Package settings provides intermediary access to settings of a Betula instance. It handles DB interaction for you. Make sure access to it is initialized.
package settings

import (
	"database/sql"
	_ "embed"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"html"
	"html/template"
	"log"
	"net/url"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/types"
)

const defaultHost = "0.0.0.0"
const biggestPort = 65535
const defaultPort = 1738

var cache types.Settings
var adminUsername string

// Those that did not fit in cache go in their own variables below. Handle with thought.
var cacheSiteDescription template.HTML

// If the port is ok, return it. Otherwise, return the default port.
func validatePortFromDB(port sql.NullInt64) uint {
	if port.Valid && port.Int64 > 0 && port.Int64 <= biggestPort {
		return uint(port.Int64)
	}

	if port.Valid && db.BookmarkCount(true) > 0 {
		log.Printf("An invalid network port is provided: %d. Using %d instead.\n", port.Int64, defaultPort)
	}

	return defaultPort
}

func validateHostFromDB(addr sql.NullString) string {
	if addr.Valid && addr.String != "" {
		return addr.String
	}

	if addr.Valid && db.BookmarkCount(true) > 0 {
		log.Printf("An invalid network host is provided: %s. Using %s instead.\n", addr.String, defaultHost)
	}

	return defaultHost
}

func ValidatePortFromWeb[N ~int | uint](port N) uint {
	if port <= 0 || port > biggestPort {
		log.Printf("An invalid network port is provided: %d. Using %d instead.\n", port, defaultPort)
		return defaultPort
	}
	return uint(port)
}

// Index reads all settings from the db.
func Index() {
	adminUsername = db.MetaEntry[sql.NullString](db.BetulaMetaAdminUsername).String

	unvalidatedNetworkHost := db.MetaEntry[sql.NullString](db.BetulaMetaNetworkHost)
	cache.NetworkHost = validateHostFromDB(unvalidatedNetworkHost)

	unvalidatedNetworkPort := db.MetaEntry[sql.NullInt64](db.BetulaMetaNetworkPort)
	cache.NetworkPort = validatePortFromDB(unvalidatedNetworkPort)

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

	// In most cases, you would need the sql.Null* types to handle
	// the case when there is no entry for the setting. For these
	// particular settings, we are perfectly fine with it just
	// returning "" when it is not present.
	cache.CustomCSS = db.MetaEntry[string](db.BetulaMetaCustomCSS)
	cache.PublicCustomJS = db.MetaEntry[string](db.BetulaMetaPublicCustomJS)
	cache.PrivateCustomJS = db.MetaEntry[string](db.BetulaMetaPrivateCustomJS)

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

	enableFederation := db.MetaEntry[sql.NullInt64](db.BetulaMetaEnableFederation)
	if !enableFederation.Valid || enableFederation.Int64 != 0 {
		cache.FederationEnabled = true
	} else {
		cache.FederationEnabled = false
	}
}

func AdminUsername() string              { return adminUsername }
func SiteURL() string                    { return cache.SiteURL }
func NetworkPort() uint                  { return cache.NetworkPort }
func NetworkHost() string                { return cache.NetworkHost }
func SiteName() string                   { return cache.SiteName }
func SiteTitle() template.HTML           { return cache.SiteTitle }
func SiteDescriptionHTML() template.HTML { return cacheSiteDescription }
func SiteDescriptionMycomarkup() string  { return cache.SiteDescriptionMycomarkup }
func CustomCSS() string                  { return cache.CustomCSS }
func FederationEnabled() bool            { return cache.FederationEnabled }
func PublicCustomJS() string             { return cache.PublicCustomJS }
func PrivateCustomJS() string            { return cache.PrivateCustomJS }

func SiteDomain() string {
	if SiteURL() == "" {
		return ""
	}
	return stricks.ParseValidURL(SiteURL()).Host
}

func UserAgent() string {
	return fmt.Sprintf("Betula; %s; Bot", SiteDomain())
}

func SetSettings(settings types.Settings) {
	if settings.SiteName == "" {
		settings.SiteName = "Betula"
	}
	db.SetMetaEntry(db.BetulaMetaNetworkHost, settings.NetworkHost)
	db.SetMetaEntry(db.BetulaMetaNetworkPort, ValidatePortFromWeb(settings.NetworkPort))
	db.SetMetaEntry(db.BetulaMetaSiteName, settings.SiteName)
	db.SetMetaEntry(db.BetulaMetaSiteTitle, string(settings.SiteTitle))
	db.SetMetaEntry(db.BetulaMetaSiteDescription, settings.SiteDescriptionMycomarkup)
	db.SetMetaEntry(db.BetulaMetaSiteURL, settings.SiteURL)
	db.SetMetaEntry(db.BetulaMetaCustomCSS, settings.CustomCSS)
	db.SetMetaEntry(db.BetulaMetaEnableFederation, settings.FederationEnabled)
	db.SetMetaEntry(db.BetulaMetaPublicCustomJS, settings.PublicCustomJS)
	db.SetMetaEntry(db.BetulaMetaPrivateCustomJS, settings.PrivateCustomJS)
	Index()
}

func WritePort(port uint) { // port must != 0
	db.SetMetaEntry(db.BetulaMetaNetworkPort, port)
	Index()
}
