package badger

import "github.com/dgraph-io/badger/v3"

func SetQueue(key string) error {
	return DB.Update(func(txn *badger.Txn) error {
		namespacedKey := []byte(":::" + key)
		return txn.Set(namespacedKey, []byte{1})
	})
}

func GetQueue(key string) bool {
	var found bool
	_ = DB.View(func(txn *badger.Txn) error {
		namespacedKey := []byte(":::" + key)
		_, err := txn.Get(namespacedKey)
		if err == nil {
			found = true
		}
		return nil
	})
	return found
}
