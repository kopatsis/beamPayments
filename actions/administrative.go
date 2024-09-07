package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/redis"

	"github.com/gobuffalo/buffalo"
)

func HandleLogAllOut(c buffalo.Context) error {
	uid, err := firebaseApp.VerifyTokenAndReturnUID(c)
	if err != nil {
		return c.Error(400, err)
	}

	err = redis.AddResetDate(uid)
	if err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func HandleDeleteAccount(c buffalo.Context) error {
	uid, err := firebaseApp.VerifyTokenAndReturnUID(c)
	if err != nil {
		return c.Error(400, err)
	}

	err = redis.AddBanned(uid)
	if err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func HandleUserLogout(c buffalo.Context) error {
	c.Session().Clear()
	err := c.Session().Save()

	response := map[string]any{"loggedout": true}
	if err != nil {
		response = map[string]any{"loggedout": false, "error": err.Error()}
	}

	return c.Render(200, r.JSON(response))
}
