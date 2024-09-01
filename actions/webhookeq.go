package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/models"
	"beam_payments/models/badger"
	"context"
	"errors"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72/sub"
)

func HandleEquivalentWebhook(c buffalo.Context) error {

	id := c.Param("id")

	subscription, err := sub.Get(id, nil)
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

}
