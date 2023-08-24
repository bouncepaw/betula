// Package db encapsulates all used queries to the database.
//
// Do not forget to Initialize and Finalize.
//
// All functions in this package might crash vividly.
package db

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"log"
)

func init() {
	var sqlite3conn []*sqlite3.SQLiteConn
	sql.Register("betulizator",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				sqlite3conn = append(sqlite3conn, conn)
				conn.RegisterUpdateHook(func(op int, db string, table string, rowid int64) {
					switch op {
					case sqlite3.SQLITE_INSERT:
						if table == "Jobs" {
							JobChannel <- rowid
						}
					}
				})
				return nil
			},
		})
}

// Initialize opens a SQLite3 database with the given filename. The connection is encapsulated, you cannot access the database directly, you are to use the functions provided by the package.
func Initialize(filename string) {
	var err error

	db, err = sql.Open("betulizator", filename+"?cache=shared")
	if err != nil {
		log.Fatalln(err)
	}

	db.SetMaxOpenConns(1)
	handleMigrations()
}

// Finalize closes the connection with the database.
func Finalize() {
	err := db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

var (
	db         *sql.DB
	JobChannel = make(chan int64)
)

// Utility functions

func mustExec(query string, args ...any) {
	_, err := db.Exec(query, args...)
	if err != nil {
		log.Fatalln(err)
	}
}

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
	var res T
	for rows.Next() { // Do 0 or 1 times
		mustScan(rows, &res)
		break
	}
	_ = rows.Close()
	return res
}
