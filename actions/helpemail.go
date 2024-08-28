package actions

import (
	"beam_payments/actions/cloudflare"
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type ContactForm struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Captcha string `json:"cf-turnstile-response"`
}

func ExternalEmailHandler(c buffalo.Context) error {
	return actualEmailFunction(c)
}

func InternalEmailHandler(c buffalo.Context) error {
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

	if err := sendFormSubmissionEmail(form.Email, form.Subject, form.Body); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"message": "Sucessfully sent the email! Expect a reply in 1-3 business days, but usually way sooner."}
	return c.Render(200, r.JSON(response))
}

func sendFormSubmissionEmail(formEmail, formSubject, formBody string) error {
	from := mail.NewEmail("No Reply", "donotreply@shortentrack.com")
	to := mail.NewEmail("Admin", "admin@shortentrack.com")
	replyTo := mail.NewEmail("", formEmail)
	subject := "FORM SUBMISSION: " + formSubject
	content := mail.NewContent("text/plain", formBody)

	message := mail.NewV3MailInit(from, subject, to, content)
	message.SetReplyTo(replyTo)

	response, err := sendgrid.SGClient.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	return nil
}
