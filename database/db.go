package database

import (
	"log"

	"github.com/dgraph-io/badger/v4"
)

var badgerDb *badger.DB

func GetDb() *badger.DB {
	return badgerDb
}

func GetDatabaseConnection() {
	db, err := badger.Open(badger.DefaultOptions("store"))
	if err != nil {
		log.Fatalln("Failed to open DB connection. Exiting ...")
	}
	defer db.Close()

	badgerDb = db
}
