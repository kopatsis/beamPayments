package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/redis"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
)

func HandleStripeWebhook(c buffalo.Context) error {
	const MaxBodyBytes = int64(65536)
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.Error(400, err)
	}

	sigHeader := c.Request().Header.Get("Stripe-Signature")
	endpointSecret := os.Getenv("END_SECR")

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		return c.Error(400, err)
	}

	switch event.Type {
	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return c.Error(400, err)
		}

		subscription, err := sub.Get(invoice.Subscription.ID, nil)
		if err != nil {
			return c.Error(400, err)
		}

		userid, err := redis.GetUserBySubID(subscription.ID)
		if err != nil {
			return c.Error(400, err)
		}

		firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
		if err != nil {
			return c.Error(400, err)
		}

		userPayment, err := redis.GetUserPayment(userid)
		if err != nil {
			return c.Error(400, err)
		} else if userPayment == nil {
			return c.Error(400, errors.New("no user payment"))
		}

		new := !userPayment.LastDate.IsZero()

		if err := redis.SetUserPaymentActive(userid, subscription.ID, time.Unix(subscription.CurrentPeriodEnd, 0)); err != nil {
			return c.Error(400, err)
		}

		if err := redis.RDB.Publish(context.Background(), "Subscriptions", subscription.ID+" --- "+"Success").Err(); err != nil {
			return c.Error(400, err)
		}

		if err := sendgrid.SendSuccessEmail(firebaseUser.Email, new); err != nil {
			return c.Error(400, errors.New("didn't send email but everything else worked: "+err.Error()))
		}

		response := map[string]any{"success": true}
		return c.Render(200, r.JSON(response))

	case "invoice.payment_failed":

		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return c.Error(400, err)
		}

		subscription, err := sub.Get(invoice.Subscription.ID, nil)
		if err != nil {
			return c.Error(400, err)
		}

		userid, err := redis.GetUserBySubID(subscription.ID)
		if err != nil {
			return c.Error(400, err)
		}

		firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
		if err != nil {
			return c.Error(400, err)
		}

		userPayment, err := redis.GetUserPayment(userid)
		if err != nil {
			return c.Error(400, err)
		} else if userPayment == nil {
			return c.Error(400, errors.New("no user payment"))
		}

		if err := redis.SetUserPaymentInactive(userid, subscription.ID); err != nil {
			return c.Error(400, err)
		}

		if err := redis.RDB.Publish(context.Background(), "Subscriptions", subscription.ID+" --- "+"Fail").Err(); err != nil {
			return c.Error(400, err)
		}

		if err := sendgrid.SendFailureEmail(firebaseUser.Email); err != nil {
			return c.Error(400, errors.New("didn't send email but everything else worked: "+err.Error()))
		}

		response := map[string]any{"success": true}
		return c.Render(200, r.JSON(response))

	case "customer.subscription.created":

	case "customer.subscription.updated":

	case "customer.subscription.deleted":

	case "charge.refunded":

	case "charge.dispute.created":

		chargeID, ok := event.Data.Object["charge"].(string)
		if !ok {
			return c.Error(400, errors.New("unable to extract charge ID from event"))
		}

		chargeObj, err := charge.Get(chargeID, nil)
		if err != nil {
			return c.Error(400, err)
		}

		invoiceID := chargeObj.Invoice.ID
		if invoiceID == "" {
			return c.Error(400, errors.New("no invoice associated with this charge"))
		}

		invoiceObj, err := invoice.Get(invoiceID, nil)
		if err != nil {
			return c.Error(400, err)
		}

		subscriptionID := invoiceObj.Subscription.ID

		subscription, err := sub.Get(subscriptionID, nil)
		if err != nil {
			return c.Error(400, err)
		}

		userid, err := redis.GetUserBySubID(subscription.ID)
		if err != nil {
			return c.Error(400, err)
		}

		email := "NO EMAIL FOR THIS ONE"

		firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
		if err == nil && firebaseUser.Email != "" {
			email = firebaseUser.Email
		}

		if err := sendgrid.SendChargeBackAlert(subscriptionID, userid, email, string(subscription.Status)); err != nil {
			return c.Error(400, err)
		}

		response := map[string]any{"success": true}
		return c.Render(200, r.JSON(response))

	}

	return c.Render(200, r.JSON(map[string]string{"status": "event received"}))
}
