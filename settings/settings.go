// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2023 ninedraft
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package settings provides intermediary access to settings of a Betula instance. It handles DB interaction for you. Make sure access to it is initialized.
package settings

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/url"
	"os"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	"git.sr.ht/~bouncepaw/betula/ports/settings"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	localBookmarks = db.NewLocalBookmarksRepo()
	settingsRepo   = &db.SettingsRepo{}
)

func mustRead[T any](v T, err error) T {
	if err != nil {
		slog.Error("Failed to read setting", "err", err)
		os.Exit(1)
	}
	return v
}

func mustWrite(err error) {
	if err != nil {
		slog.Error("Failed to write setting", "err", err)
		os.Exit(1)
	}
}

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

	count, _ := localBookmarks.BookmarkCount(context.Background(), true)
	if port.Valid && count > 0 {
		slog.Warn("Invalid network port from DB, using default", "port", port.Int64, "default", defaultPort)
	}

	return defaultPort
}

func validateHostFromDB(addr sql.NullString) string {
	if addr.Valid && addr.String != "" {
		return addr.String
	}

	count, _ := localBookmarks.BookmarkCount(context.Background(), true)
	if addr.Valid && count > 0 {
		slog.Warn("Invalid network host from DB, using default", "host", addr.String, "default", defaultHost)
	}

	return defaultHost
}

func ValidatePortFromWeb[N ~int | uint](port N) uint {
	if port <= 0 || port > biggestPort {
		slog.Warn("Invalid network port from web, using default", "port", port, "default", defaultPort)
		return defaultPort
	}
	return uint(port)
}

// Index reads all settings from the db.
func Index() {
	ctx := context.Background()
	adminUsername = mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaAdminUsername)).String

	unvalidatedNetworkHost := mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaNetworkHost))
	cache.NetworkHost = validateHostFromDB(unvalidatedNetworkHost)

	unvalidatedNetworkPort := mustRead(settingsRepo.MetaEntryNullInt64(ctx, settingsports.BetulaMetaNetworkPort))
	cache.NetworkPort = validatePortFromDB(unvalidatedNetworkPort)

	siteName := mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaSiteName))
	if siteName.Valid && siteName.String != "" {
		cache.SiteName = siteName.String
	} else {
		cache.SiteName = "Betula"
	}

	siteTitle := mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaSiteTitle))
	if siteTitle.Valid && siteTitle.String != "" {
		cache.SiteTitle = template.HTML(siteTitle.String)
	} else {
		cache.SiteTitle = template.HTML(html.EscapeString(cache.SiteName))
	}

	siteDescription := mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaSiteDescription))
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
	cache.CustomCSS = mustRead(settingsRepo.MetaEntryString(ctx, settingsports.BetulaMetaCustomCSS))
	cache.PublicCustomJS = mustRead(settingsRepo.MetaEntryString(ctx, settingsports.BetulaMetaPublicCustomJS))
	cache.PrivateCustomJS = mustRead(settingsRepo.MetaEntryString(ctx, settingsports.BetulaMetaPrivateCustomJS))

	siteURL := mustRead(settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaSiteURL))
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

	enableFederation := mustRead(settingsRepo.MetaEntryNullInt64(ctx, settingsports.BetulaMetaEnableFederation))
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
	return bxstr.ParseValidURL(SiteURL()).Host
}

func UserAgent() string {
	return fmt.Sprintf("Betula; %s; Bot", SiteDomain())
}

func SetSettings(settings types.Settings) {
	if settings.SiteName == "" {
		settings.SiteName = "Betula"
	}
	ctx := context.Background()
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaNetworkHost, settings.NetworkHost))
	mustWrite(settingsRepo.SetMetaEntryUint(ctx, settingsports.BetulaMetaNetworkPort, ValidatePortFromWeb(settings.NetworkPort)))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaSiteName, settings.SiteName))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaSiteTitle, string(settings.SiteTitle)))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaSiteDescription, settings.SiteDescriptionMycomarkup))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaSiteURL, settings.SiteURL))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaCustomCSS, settings.CustomCSS))
	mustWrite(settingsRepo.SetMetaEntryBool(ctx, settingsports.BetulaMetaEnableFederation, settings.FederationEnabled))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaPublicCustomJS, settings.PublicCustomJS))
	mustWrite(settingsRepo.SetMetaEntryString(ctx, settingsports.BetulaMetaPrivateCustomJS, settings.PrivateCustomJS))
	Index()
}

func WritePort(port uint) { // port must != 0
	mustWrite(settingsRepo.SetMetaEntryUint(context.Background(), settingsports.BetulaMetaNetworkPort, port))
	Index()
}
