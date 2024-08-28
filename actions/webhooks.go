package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/models"
	"beam_payments/models/badger"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
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

		userid, err := models.ConfirmSubscription(subscription.ID, time.Unix(subscription.CurrentPeriodEnd, 0))
		if err != nil {
			return c.Error(400, err)
		}

		firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
		if err != nil {
			return c.Error(400, err)
		}

		isFirst := !badger.GetQueue(subscription.ID)

		if err := badger.SetQueue(subscription.ID); err != nil {
			return c.Error(400, err)
		}

		if err := sendgrid.SendSuccessEmail(firebaseUser.Email, isFirst); err != nil {
			return c.Error(400, errors.New("didn't send email but everything else worked: "+err.Error()))
		}

		response := map[string]any{"success": true}
		return c.Render(200, r.JSON(response))

	case "invoice.payment_failed":

		// Email user payment failed and will lose payinng in 3 days (72 hours) if not. Even if after, can re-activate with payment

	case "customer.subscription.created":

	case "customer.subscription.updated":

	case "customer.subscription.deleted":

	case "charge.refunded":

	case "charge.dispute.created":

		// Email ME personally

	}

	return c.Render(200, r.JSON(map[string]string{"status": "event received"}))
}
