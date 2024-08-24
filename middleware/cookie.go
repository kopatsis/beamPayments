package middleware

import (
	"beam_payments/models/badger"
	"errors"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
)

func CookieMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {

		path := c.Request().URL.Path

		if path == "/multipass" || strings.HasPrefix(path, "/webhook") {
			return next(c)
		}

		userID := c.Session().Get("user_id").(string)
		passcodeStr := c.Session().Get("passcode").(string)
		dateStr := c.Session().Get("date").(string)

		if userID == "" || passcodeStr == "" || dateStr == "" {
			return c.Error(401, errors.New("unauthorized: missing session data"))
		}

		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return c.Error(401, errors.New("unauthorized: missing session data"))
		}

		authorized, banned := badger.CheckCookie(userID, passcodeStr, date)

		if !authorized {
			return c.Error(400, errors.New("unauthorized: not logged in"))
		}

		if banned {
			c.Session().Clear()
			c.Session().Save()
			return c.Error(400, errors.New("unauthorized: user does not exist"))
		}

		return next(c)
	}
}
