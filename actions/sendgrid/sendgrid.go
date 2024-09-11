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

func SendFailureEmail(userEmail string) error {
	from := mail.NewEmail("Shorten Track Team", "donotreply@shortentrack.com")
	to := mail.NewEmail(strings.Split(userEmail, "@")[0], userEmail)

	subject := "Your Payment Method Failed for Your Shorten Track Membership"

	content := mail.NewContent("text/plain", "We noticed an issue with your payment for Shorten Track.\n\nUnfortunately, we were unable to process your payment for this month's membership. Please update your payment information at pay.shortentrack.com to continue enjoying our services.\n\nBest regards,\nThe Shorten Track Team")

	htmlContent := mail.NewContent("text/html", "<p>We noticed an issue with your payment for <strong>Shorten Track.</strong></p><p>Unfortunately, we were unable to process your payment for this month's membership. Please update your payment information at <a href='https://pay.shortentrack.com'>pay.shortentrack.com</a> to continue enjoying our services.</p><p>Best regards,<br>The Shorten Track Team</p>")

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

func SendCancelEmail(userEmail string, isCancelled bool) error {
	from := mail.NewEmail("Shorten Track Team", "donotreply@shortentrack.com")
	to := mail.NewEmail(strings.Split(userEmail, "@")[0], userEmail)

	var subject, plainTextContent, htmlTextContent string

	if isCancelled {
		subject = "Your Shorten Track Membership Has Been Cancelled"
		plainTextContent = "We're sorry to see you go. Your Shorten Track membership has been cancelled. If this was a mistake or you wish to rejoin, please update your membership information at pay.shortentrack.com.\n\nBest regards,\nThe Shorten Track Team"
		htmlTextContent = "<p>We're sorry to see you go. Your <strong>Shorten Track</strong> membership has been cancelled. If this was a mistake or you wish to rejoin, please update your membership information at <a href='https://pay.shortentrack.com'>pay.shortentrack.com</a>.</p><p>Best regards,<br>The Shorten Track Team</p>"
	} else {
		subject = "Your Shorten Track Membership Has Been Reactivated"
		plainTextContent = "Welcome back! Your Shorten Track membership has been reactivated. You can manage your membership anytime at pay.shortentrack.com.\n\nBest regards,\nThe Shorten Track Team"
		htmlTextContent = "<p>Welcome back! Your <strong>Shorten Track</strong> membership has been reactivated. You can manage your membership anytime at <a href='https://pay.shortentrack.com'>pay.shortentrack.com</a>.</p><p>Best regards,<br>The Shorten Track Team</p>"
	}

	content := mail.NewContent("text/plain", plainTextContent)
	htmlContent := mail.NewContent("text/html", htmlTextContent)

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

func SendPaymentUpdateEmail(userEmail string) error {
	from := mail.NewEmail("Shorten Track Team", "donotreply@shortentrack.com")
	to := mail.NewEmail(strings.Split(userEmail, "@")[0], userEmail)

	subject := "Your Default Payment Information Has Been Updated"
	plainTextContent := "We wanted to let you know that your default payment information for Shorten Track has been successfully updated. You can manage your payment methods at pay.shortentrack.com.\n\nBest regards,\nThe Shorten Track Team"
	htmlTextContent := "<p>We wanted to let you know that your default payment information for <strong>Shorten Track</strong> has been successfully updated. You can manage your payment methods at <a href='https://pay.shortentrack.com'>pay.shortentrack.com</a>.</p><p>Best regards,<br>The Shorten Track Team</p>"

	content := mail.NewContent("text/plain", plainTextContent)
	htmlContent := mail.NewContent("text/html", htmlTextContent)

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

func SendSeriousErrorAlert(iss, body string) error {
	from := mail.NewEmail("No Reply", "donotreply@shortentrack.com")
	to := mail.NewEmail("Admin", "admin@shortentrack.com")
	subject := "ALERT ISSUE WITHIN APPLICATION: " + iss

	content := mail.NewContent("text/plain", body)

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
