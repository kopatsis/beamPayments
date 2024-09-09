package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/middleware"
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
	middleware.RemoveCookie(c)

	response := map[string]any{"loggedout": true}
	return c.Render(200, r.JSON(response))
}
