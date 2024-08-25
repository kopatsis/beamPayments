package multipass

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/models/badger"
	"net/http"
	"os"
	"time"

	"github.com/gobuffalo/buffalo"
)

func Multipass(c buffalo.Context) error {

	originalURL := os.Getenv("OG_URL")
	if originalURL == "" {
		originalURL = "https://shortentrack.com"
	}

	uid, err := firebaseApp.VerifyTokenAndReturnUID(c)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, originalURL+"/login?red=t")
	}

	passcode, banned, err := badger.CreateCookie(uid)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/loginerror")
	}

	if banned {
		return c.Redirect(http.StatusSeeOther, "/loginerror")
	}

	c.Session().Set("user_id", uid)
	c.Session().Set("passcode", passcode)
	c.Session().Set("date", time.Now().Format(time.RFC3339))

	if err := c.Session().Save(); err != nil {
		return c.Redirect(http.StatusSeeOther, "/loginerror")
	}

	return c.Redirect(http.StatusSeeOther, "/")

}
