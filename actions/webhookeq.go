package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
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

}
