// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

// Command betula runs Betula, a personal link collection software.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	apgw "git.sr.ht/~bouncepaw/betula/gateways/activitypub"
	wwwgw "git.sr.ht/~bouncepaw/betula/gateways/www"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/settings"
	archivingsvc "git.sr.ht/~bouncepaw/betula/svc/archiving"
	feedssvc "git.sr.ht/~bouncepaw/betula/svc/feeds"
	helpingsvc "git.sr.ht/~bouncepaw/betula/svc/helping"
	likingsvc "git.sr.ht/~bouncepaw/betula/svc/liking"
	notifsvc "git.sr.ht/~bouncepaw/betula/svc/notif"
	remarkingsvc "git.sr.ht/~bouncepaw/betula/svc/remarking"
	searchsvc "git.sr.ht/~bouncepaw/betula/svc/searching"
	settingssvc "git.sr.ht/~bouncepaw/betula/svc/settings"
	"git.sr.ht/~bouncepaw/betula/web"
	_ "git.sr.ht/~bouncepaw/betula/web" // For init()
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	var port uint
	var versionFlag bool

	flag.BoolVar(&versionFlag, "version", false, "Print version and exit.")
	flag.UintVar(&port, "port", 0, "Port number. "+
		"The value gets written to a database file and is used immediately.")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage: %s DB_PATH.betula\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if versionFlag {
		fmt.Printf("Betula %s\n", "v1.7.0")
		return
	}

	if len(flag.Args()) < 1 {
		slog.Error("Pass a database file name!")
		os.Exit(1)
	}

	filename, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		slog.Error("Failed to resolve database path", "err", err)
		os.Exit(1)
	}

	fmt.Println("Hello Betula!")

	db.Initialize(filename)
	defer db.Finalize()
	settings.Index()
	auth.Initialize()
	// If the user provided a non-zero port, use it. Write it to the DB. It will be picked up later by settings.Index(). If they did not provide such a port, whatever, settings.Index() will figure something out 🙏
	if port > 0 {
		settings.WritePort(port)
	}
	signing.EnsureKeysFromDatabase()
	activities.GenerateBetulaActor()
	go jobs.ListenAndWhisper()
	web.StartServer(newController())
}

func newController() web.Controller {
	var (
		repoLike           = db.NewLikeRepo()
		repoLikeCollection = db.NewLikeCollectionRepo()
		repoNotif          = db.New()
		repoActor          = db.NewActorRepo()
		repoLocalBookmark  = db.NewLocalBookmarksRepo()
		repoRemoteBookmark = db.NewRemoteBookmarkRepo()
		repoArchives       = db.NewArchivesRepo()
		repoSettings       = &db.SettingsRepo{}

		obeliskFetcher = archivingsvc.NewObeliskFetcher()
		activityPub    = apgw.NewActivityPub(repoActor, repoRemoteBookmark)
		www            = wwwgw.New()

		// One day, all shall be in services!
		svcSettings  = settingssvc.New(repoSettings, "v1.7.0", settings.SiteDomain)
		svcNotif     = notifsvc.New(repoNotif)
		svcArchiving = archivingsvc.New(obeliskFetcher, repoArchives)
		svcLiking    = likingsvc.New(
			repoLike,
			repoLikeCollection,
			repoLocalBookmark,
			repoNotif,
			activityPub)
		svcRemarking = remarkingsvc.New(activityPub)
		svcFeeds     = feedssvc.New()
		svcSearching = searchsvc.New()
		svcHelping   = helpingsvc.New()
	)

	return web.Controller{
		SvcNotif:           svcNotif,
		SvcArchiving:       svcArchiving,
		SvcLiking:          svcLiking,
		SvcRemarking:       svcRemarking,
		SvcFeeds:           svcFeeds,
		SvcSearching:       svcSearching,
		SvcHelping:         svcHelping,
		SvcSettings:        svcSettings,
		ActivityPub:        activityPub,
		WWW:                www,
		RepoRemoteBookmark: repoRemoteBookmark,
		RepoActor:          repoActor,
	}
}
