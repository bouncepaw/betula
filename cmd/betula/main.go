// Command betula runs Betula, a personal link collection software.
package main

import (
	"flag"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/web"
	_ "git.sr.ht/~bouncepaw/betula/web" // For init()
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Hello Betula!")

	var port uint

	flag.UintVar(&port, "port", 0, "port number. "+
		"The value gets written to a database file.")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalln("Pass a database file name!")
	}

	filename, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	db.Initialize(filename)
	defer db.Finalize()
	auth.Initialize()
	// If user didn't provide the port, check the port from database
	if port == 0 {
		dbPort := db.MetaEntry[uint](db.BetulaMetaNetworkPort)
		settings.SetNetworkPort(settings.Uintport(dbPort).ValidatePort())
	} else {
		// Check the user provided port
		settings.SetNetworkPort(settings.Uintport(port).ValidatePort())
	}
	settings.Index()
	web.StartServer()
}
