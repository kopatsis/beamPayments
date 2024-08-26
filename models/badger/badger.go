package badger

import (
	"log"

	"github.com/dgraph-io/badger/v3"
)

var DB *badger.DB

func init() {
	opts := badger.DefaultOptions("./badgerdb")
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open BadgerDB: %v", err)
	}

	DB = db
}

func Close() {
	if err := DB.Close(); err != nil {
		log.Fatalf("Failed to close BadgerDB: %v", err)
	}
}
