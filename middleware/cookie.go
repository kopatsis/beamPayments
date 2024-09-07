package middleware

import (
	"beam_payments/models/badger"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
)

func errorSplit(c buffalo.Context, err error, banned bool) error {
	method := c.Request().Method
	if method == "GET" {
		if !banned {
			loginURL := os.Getenv("OG_URL")
			if loginURL == "" {
				loginURL = "https://shortentrack.com"
			}
			return c.Redirect(http.StatusSeeOther, loginURL+"/login?circleRedir=t")
		} else {
			return c.Redirect(http.StatusSeeOther, "/loginerror")
		}

	}
	return c.Error(401, err)
}

func CookieMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {

		fmt.Println("faggot")

		path := c.Request().URL.Path

		if path == "/multipass" || strings.HasPrefix(path, "/webhook") || strings.HasPrefix(path, "/administrative") || strings.HasPrefix(path, "/check") {
			return next(c)
		}

		userID := c.Session().Get("user_id").(string)
		passcodeStr := c.Session().Get("passcode").(string)
		dateStr := c.Session().Get("date").(string)

		if userID == "" || passcodeStr == "" || dateStr == "" {
			return errorSplit(c, errors.New("unauthorized: missing session data"), false)
		}

		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return errorSplit(c, errors.New("unauthorized: missing session data"), false)
		}

		authorized, banned := badger.CheckCookie(userID, passcodeStr, date)

		if !authorized {
			return errorSplit(c, errors.New("unauthorized: not logged in"), false)
		}

		if banned {
			c.Session().Clear()
			c.Session().Save()
			return errorSplit(c, errors.New("unauthorized: user does not exist"), true)
		}

		return next(c)
	}
}
