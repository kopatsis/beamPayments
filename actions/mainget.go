package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/stripefunc"
	"beam_payments/models"
	"context"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/setupintent"
	"github.com/stripe/stripe-go/v72/sub"
)

func GetHandler(c buffalo.Context) error {
	userID := c.Session().Get("user_id").(string)
	if userID == "" {
		return c.Redirect(http.StatusSeeOther, "/error")
	}

	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userID)
	if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
		return c.Redirect(http.StatusSeeOther, "/error")
	}

	dbsub, exists, err := models.GetSubscription(userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/error")
	}
	c.Set("Email", firebaseUser.Email)

	if !exists || dbsub == nil {
		params := &stripe.SetupIntentParams{
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
		}
		si, err := setupintent.New(params)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/error")
		}
		c.Set("Secret", si.ClientSecret)

		return c.Render(http.StatusOK, r.HTML("all/pay.plush.html"))
	} else {

		s, err := sub.Get(dbsub.SubscriptionID, nil)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/error")
		}

		if dbsub.Processing {
			return c.Render(http.StatusOK, r.HTML("all/processing.plush.html"))
		}

		if dbsub.Ending || s.CancelAtPeriodEnd {
			c.Set("REndDate", time.Unix(s.CurrentPeriodEnd, 0))
			c.Set("EndDate", dbsub.EndDate)
			return c.Render(http.StatusOK, r.HTML("all/ending.plush.html"))
		}

		paymentType, cardBrand, lastFour, err := stripefunc.GetPaymentMethodDetails(s.ID)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/error")
		}

		params := &stripe.SetupIntentParams{
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
		}
		si, err := setupintent.New(params)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/error")
		}

		expiring := false
		if dbsub.ExpiresDate.Before(time.Now()) {
			expiring = true
		}

		c.Set("PaymentType", paymentType)
		c.Set("CardBrand", cardBrand)
		c.Set("LastFour", lastFour)
		c.Set("Expiring", expiring)
		c.Set("Secret", si.ClientSecret)
		c.Set("REndDate", time.Unix(s.CurrentPeriodEnd, 0))
		c.Set("EndDate", dbsub.ExpiresDate)
		return c.Render(http.StatusOK, r.HTML("all/admin.plush.html"))
	}
}
