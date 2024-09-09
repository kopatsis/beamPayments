package middleware

import (
	"beam_payments/redis"
	"errors"
	"net/http"
	"os"
	"strings"

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

		path := c.Request().URL.Path

		if path == "/multipass" || strings.HasPrefix(path, "/webhook") || strings.HasPrefix(path, "/administrative") || strings.HasPrefix(path, "/check") {
			return next(c)
		}

		userID, date, err := GetCookie(c)
		if err != nil {
			return errorSplit(c, errors.New("unauthorized: missing cookie data"), false)
		}

		ban, reset, err := redis.CheckCookeLimit(userID, date)
		if err != nil {
			return errorSplit(c, errors.New("unauthorized: failure with redis"), false)
		}

		if reset {
			return errorSplit(c, errors.New("unauthorized: not logged in"), false)
		}

		if ban {
			RemoveCookie(c)
			return errorSplit(c, errors.New("unauthorized: user does not exist"), true)
		}

		return next(c)
	}
}
