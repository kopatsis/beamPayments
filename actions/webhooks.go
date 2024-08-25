package actions

import (
	"encoding/json"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
)

func HandleStripeWebhook(c buffalo.Context) error {
	body := c.Request().Body
	event := &stripe.Event{}

	if err := json.NewDecoder(body).Decode(event); err != nil {
		return c.Error(400, err)
	}

	switch event.Type {
	case "invoice.payment_succeeded":

	case "invoice.payment_failed":

	case "customer.subscription.created":

	case "customer.subscription.updated":

	case "customer.subscription.deleted":

	case "charge.refunded":

	case "charge.dispute.created":

	default:

	}

	return c.Render(200, r.JSON(map[string]string{"status": "event received"}))
}
