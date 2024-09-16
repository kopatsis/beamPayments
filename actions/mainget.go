package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/stripefunc"
	"beam_payments/middleware"
	"beam_payments/redis"
	"context"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/setupintent"
	"github.com/stripe/stripe-go/v72/sub"
)

func GetHandler(c buffalo.Context) error {

	userID, err := middleware.GetCookieUserID(c)
	if err != nil {
		c.Set("Error", "No user id in cookie available: "+err.Error())
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userID)
	if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
		if err != nil {
			c.Set("Error", err.Error())
		} else {
			c.Set("Error", "Email not verified or no email at all.")
		}
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	c.Set("Email", firebaseUser.Email)
	userPayment, err := redis.GetUserPayment(userID)
	if err != nil {
		c.Set("Error", err.Error())
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	if userPayment == nil {
		params := &stripe.SetupIntentParams{
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
		}
		si, err := setupintent.New(params)
		if err != nil {
			c.Set("Error", err.Error())
			return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
		}
		c.Set("Secret", si.ClientSecret)

		return c.Render(http.StatusOK, r.HTML("all/pay.plush.html"))
	}

	s, err := sub.Get(userPayment.SubscriptionID, nil)
	if err != nil {
		c.Set("Error", err.Error())
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	switch string(s.Status) {
	case "incomplete":
		c.Set("ID", s.ID)
		return c.Render(http.StatusOK, r.HTML("all/processing.plush.html"))
	case "past_due":
		c.Set("ID", s.ID)
		return c.Render(http.StatusOK, r.HTML("all/updatepay.plush.html"))
	}

	if s.CancelAtPeriodEnd {
		c.Set("EndDate", time.Unix(s.CurrentPeriodEnd, 0))
		return c.Render(http.StatusOK, r.HTML("all/ending.plush.html"))
	}

	params := &stripe.SetupIntentParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}
	si, err := setupintent.New(params)
	if err != nil {
		c.Set("Error", err.Error())
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	paymentType, cardBrand, lastFour, expMonth, expYear, err := stripefunc.GetPaymentMethodDetails(s.ID)
	if err != nil {
		c.Set("Error", err.Error())
		return c.Render(http.StatusOK, r.HTML("error/error.plush.html"))
	}

	c.Set("PaymentType", paymentType)
	c.Set("CardBrand", cardBrand)
	c.Set("LastFour", lastFour)
	c.Set("ExpMonth", expMonth)
	c.Set("ExpYear", expYear)
	c.Set("Secret", si.ClientSecret)
	c.Set("EndDate", time.Unix(s.CurrentPeriodEnd, 0))
	return c.Render(http.StatusOK, r.HTML("all/admin.plush.html"))
}
