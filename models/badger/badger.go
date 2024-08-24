package badger

import (
	"encoding/json"
	"log"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofrs/uuid"
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

func CreateCookie(uid string) (string, bool, error) {

	var cookie Cookie
	exists := true

	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(uid))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				exists = false
				return nil
			}
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, &cookie)
	})

	if err != nil {
		return "", false, err
	}

	if !exists {
		newPasscode, err := uuid.NewV4()
		if err != nil {
			return "", false, err
		}

		cookie = Cookie{
			Banned:    false,
			Passcode:  newPasscode,
			ResetDate: time.Now(),
		}

		err = DB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(uid), cookie.MarshalBinary())
		})
		if err != nil {
			return "", false, err
		}

		return newPasscode.String(), false, nil
	}

	if cookie.Banned {
		return "", true, nil
	}

	return cookie.Passcode.String(), false, nil
}
