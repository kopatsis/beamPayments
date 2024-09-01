package actions

import (
	"beam_payments/models/badger"
	"net/http"
	"os"

	badg "github.com/dgraph-io/badger/v3"
	"github.com/gobuffalo/buffalo"
	"github.com/google/uuid"
)

func AddExchange(c buffalo.Context) error {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid JSON"}))
	}

	if c.Request().Header.Get("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Unauthorized"}))
	}

	id := uuid.New().String()

	err := badger.DB.Update(func(txn *badg.Txn) error {
		return txn.Set([]byte("_-_"+id), []byte(req.Email))
	})
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Failed to store data"}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]string{"id": id}))
}

func GetExchange(c buffalo.Context) error {
	if c.Request().Header.Get("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Unauthorized"}))
	}

	id := c.Param("id")

	var email string
	err := badger.DB.Update(func(txn *badg.Txn) error {
		item, err := txn.Get([]byte("_-_" + id))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			email = string(val)
			return nil
		})
		if err != nil {
			return err
		}
		return txn.Delete([]byte(id))
	})

	if err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Key not found or failed to retrieve"}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]string{"email": email}))
}
