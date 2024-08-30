package actions

import (
	"beam_payments/actions/cloudflare"
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"

	"github.com/gobuffalo/buffalo"
)

type ContactForm struct {
	Email   string `form:"email"`
	Name    string `form:"name"`
	Subject string `form:"subject"`
	Body    string `form:"body"`
	Captcha string `form:"cf-turnstile-response"`
}

func InternalEmailHandler(c buffalo.Context) error {
	return actualEmailFunction(c)
}

func ExternalEmailHandler(c buffalo.Context) error {
	_, err := firebaseApp.VerifyTokenAndReturnUID(c)
	if err != nil {
		return c.Error(400, err)
	}
	return actualEmailFunction(c)
}

func actualEmailFunction(c buffalo.Context) error {

	var form ContactForm

	if err := c.Bind(&form); err != nil {
		return c.Error(400, err)
	}

	success, err := cloudflare.VerifyTurnstile(form.Captcha)
	if err != nil {
		return c.Error(400, err)
	} else if !success {
		return c.Render(400, r.JSON(map[string]string{"message": "Unfortunately, your submission did not pass the Cloudflare verification. Close this window and try again."}))
	}

	if err := sendgrid.SendFormSubmissionEmail(form.Email, form.Name, form.Subject, form.Body); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"message": "Sucessfully sent the email! Expect a reply in 1-3 business days, but usually way sooner."}
	return c.Render(200, r.JSON(response))
}
