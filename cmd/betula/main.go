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

	port := flag.Uint("port", 1738, "port number. "+
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
	settings.SetNetworkPort(*port)
	settings.Index()
	web.StartServer()
}
