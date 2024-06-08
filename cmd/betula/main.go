// Command betula runs Betula, a personal link collection software.
package main

import (
	"flag"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/web"
	_ "git.sr.ht/~bouncepaw/betula/web" // For init()
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
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
		fmt.Printf("Betula %s\n", "v1.3.0")
		return
	}

	if len(flag.Args()) < 1 {
		log.Fatalln("Pass a database file name!")
	}

	filename, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
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
	web.StartServer()
}
