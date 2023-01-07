package db

import (
	"database/sql"
	"log"
)

// Initialize opens a SQLite3 database with the given filename. The connection is encapsulated, you cannot access the database directly, you are to use the functions provided by the package.
func Initialize(filename string) {
	var err error
	db, err = sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalln(err)
	}
}

// Finalize closes the connection with the database.
func Finalize() {
	err := db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

var (
	db *sql.DB
)

// Utility functions

func mustQuery(query string, args ...any) *sql.Rows {
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalln(err)
	}
	return rows
}

func mustScan(rows *sql.Rows, dest ...any) {
	err := rows.Scan(dest...)
	if err != nil {
		log.Fatalln(err)
	}
}

func querySingleValue[T any](query string, vals ...any) T {
	rows := mustQuery(query, vals...)
	rows.Next()
	var res T
	mustScan(rows, &res)
	_ = rows.Close()
	return res
}
