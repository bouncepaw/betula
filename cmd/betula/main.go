// Command betula runs Betula, a personal link collection software.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"git.sr.ht/~bouncepaw/betula/db"
	_ "git.sr.ht/~bouncepaw/betula/web" // For init()

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Hello Betula!")

	if len(os.Args) < 2 {
		log.Fatalln("Pass a database file name!")
	}

	filename, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	db.Initialize(filename)
	defer db.Finalize()

	// TODO: make it configurable
	err = http.ListenAndServe(":1738", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
