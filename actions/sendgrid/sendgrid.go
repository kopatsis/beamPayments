package sendgrid

import (
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SGClient *sendgrid.Client

func init() {
	SGClient = sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
}

func SendFormSubmissionEmail(formEmail, formName, formSubject, formBody string) error {
	from := mail.NewEmail("No Reply", "donotreply@shortentrack.com")
	to := mail.NewEmail("Admin", "admin@shortentrack.com")
	replyTo := mail.NewEmail("", formEmail)
	subject := "FORM SUBMISSION: " + formSubject
	content := mail.NewContent("text/plain", "NAME: "+formName+"\n"+formBody)

	message := mail.NewV3MailInit(from, subject, to, content)
	message.SetReplyTo(replyTo)

	response, err := SGClient.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	return nil
}

func SendChargeBackAlert(subID, userID, userEmail string, archived bool, dbID int) error {
	from := mail.NewEmail("No Reply", "donotreply@shortentrack.com")
	to := mail.NewEmail("Admin", "admin@shortentrack.com")
	subject := "ALERT CHARGEBACK: " + subID

	toMe := fmt.Sprintf("Someone tried to chargeback >:(\n\nDetails:\nSubscription ID: %s\nUser ID: %s\nEmail: %s\nArchived: %t\nDatabase ID: %d\n", subID, userID, userEmail, archived, dbID)

	content := mail.NewContent("text/plain", toMe)

	message := mail.NewV3MailInit(from, subject, to, content)

	response, err := SGClient.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	return nil
}

func SendSuccessEmail(userEmail string, isFirst bool) error {
	from := mail.NewEmail("Shorten Track Team", "donotreply@shortentrack.com")
	to := mail.NewEmail(strings.Split(userEmail, "@")[0], userEmail)

	var subject string
	if isFirst {
		subject = "Congrats! Your Payment Has Processed and Your Shorten Track Membership Has Started"
	} else {
		subject = "Your Payment Has Processed and Your Shorten Track Membership Will Continue"
	}

	content := mail.NewContent("text/plain", "Welcome to Shorten Track!\n\nThank you for joining our monthly membership. We're excited to have you on board.\n\nYou can manage your membership anytime at pay.shortentrack.com.\n\nBest regards,\nThe Shorten Track Team")

	htmlContent := mail.NewContent("text/html", "<p>Welcome to <strong>Shorten Track!</strong></p><p>Thank you for joining our monthly membership. We're excited to have you on board.</p><p>You can manage your membership anytime at <a href='https://pay.shortentrack.com'>pay.shortentrack.com</a>.</p><p>Best regards,<br>The Shorten Track Team</p>")

	message := mail.NewV3MailInit(from, subject, to, content)
	message.AddContent(htmlContent)

	response, err := SGClient.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: %s", response.Body)
	}

	return nil
}
