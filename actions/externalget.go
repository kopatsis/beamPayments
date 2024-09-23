package actions

import (
	"beam_payments/actions/cloudflare"
	"beam_payments/redis"
	"errors"
	"os"

	"github.com/gobuffalo/buffalo"
)

type TurnstilePost struct {
	Email   string `json:"email"`
	Captcha string `json:"cf-turnstile-response"`
}

func ExternalGetHandler(c buffalo.Context) error {

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

	id := c.Param("id")
	if id == "" {
		return c.Error(400, errors.New("no param provided"))
	}

	paying, err := redis.CheckUserPaying(id)
	if err != nil && !paying {
		return c.Error(400, errors.New("unable to query payment status"))
	}

	return c.Render(200, r.JSON(map[string]any{"paying": paying}))
}

func VerifyTurnstileHandler(c buffalo.Context) error {

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

	var form ContactForm

	if err := c.Bind(&form); err != nil {
		return c.Error(400, err)
	} else if form.Email == "" {
		return c.Render(400, r.JSON(map[string]string{"message": "Please supply an email along with the turnstile verification."}))
	}

	success, err := cloudflare.VerifyTurnstile(form.Captcha)
	if err != nil {
		return c.Error(400, err)
	} else if !success {
		return c.Render(401, r.JSON(map[string]string{"message": "Unfortunately, your submission did not pass the Cloudflare verification. Close this window and try again."}))
	}

	return c.Render(204, nil)

}
