package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/middleware"
	"beam_payments/redis"
	"errors"
	"os"

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

func HandleInternalAlertEmail(c buffalo.Context) error {

	shouldBe := os.Getenv("CHECK_PASSCODE")
	if shouldBe == "" {
		return c.Error(500, errors.New("no passcode exists on backend"))
	}

	passcode := c.Request().Header.Get("X-Passcode-ID")
	if passcode == "" {
		return c.Error(400, errors.New("no passcode provided"))
	} else if passcode != shouldBe {
		return c.Error(400, errors.New("incorrect passcode provided"))
	}

	var payload struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.Error(400, err)
	}

	err := sendgrid.SendSeriousErrorAlert(payload.Subject, payload.Body)
	if err != nil {
		sendgrid.SendSeriousErrorAlert("Sending the Actual Issue Email", "This error: "+err.Error())
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}
