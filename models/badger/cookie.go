package badger

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofrs/uuid"
)

type Cookie struct {
	Banned    bool
	Passcode  uuid.UUID
	ResetDate time.Time
}

func (c *Cookie) MarshalBinary() []byte {
	return c.Passcode.Bytes()
}

func CreateCookie(uid string) (uuidstr string, banned bool, anyerr error) {

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

func CheckCookie(uid, passcode string, created time.Time) (authorized bool, banned bool) {
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

	if err != nil || !exists {
		return false, false
	}

	if cookie.Banned {
		return false, true
	}

	if created.Before(cookie.ResetDate) {
		return false, false
	}

	if chuuid, err := uuid.FromString(passcode); err != nil {
		return false, false
	} else if chuuid != cookie.Passcode {
		return false, false
	}

	return true, false
}

func AdminCookieModify(uid string, ban bool) (bool, error) {
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
		return false, err
	}

	if !exists {
		return false, nil
	}

	if ban {
		cookie.Banned = false
	} else {
		cookie.ResetDate = time.Now()
	}

	err = DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(uid), cookie.MarshalBinary())
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
