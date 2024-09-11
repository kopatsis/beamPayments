package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/models"
	"beam_payments/redis"
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

	exists, _ := redis.GetSub(userid)
	if !exists {
		err := redis.AddSub(userid)
		if err != nil {
			return c.Error(400, err)
		}

		err = redis.RDB.Publish(context.Background(), "Subscriptions", subscription.ID+" --- "+"Success").Err()
		if err != nil {
			return c.Error(400, err)
		}
	}

	if err := sendgrid.SendSuccessEmail(firebaseUser.Email, !exists); err != nil {
		return c.Error(400, errors.New("didn't send email but everything else worked: "+err.Error()))
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))

}
