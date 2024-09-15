package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/actions/sendgrid"
	"beam_payments/middleware"
	"beam_payments/redis"
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type PaymentRequest struct {
	PaymentMethodID string `json:"paymentMethodID"`
}

func PostPayHandler(c buffalo.Context) error {

	start := time.Now()

	req := PaymentRequest{}
	if err := c.Bind(&req); err != nil {
		return c.Error(400, err)
	}

	userid, err := middleware.GetCookieUserID(c)
	if err != nil {
		return c.Error(400, errors.New("no user in :"+err.Error()))
	}

	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
	if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
		return c.Error(400, err)
	}

	email := firebaseUser.Email

	customerParams := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	customerParams.Metadata = map[string]string{
		"userId": userid,
	}

	stripeCustomer, err := customer.New(customerParams)
	if err != nil {
		return c.Error(400, err)
	}

	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(stripeCustomer.ID),
	}
	_, err = paymentmethod.Attach(req.PaymentMethodID, attachParams)
	if err != nil {
		return c.Error(400, err)
	}

	customerUpdateParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
		},
	}
	_, err = customer.Update(stripeCustomer.ID, customerUpdateParams)
	if err != nil {
		return c.Error(400, err)
	}

	priceID := os.Getenv("PRICE_ID")

	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomer.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}
	subscriptionParams.AddMetadata("userId", userid)

	newSub, err := sub.New(subscriptionParams)
	if err != nil {
		return c.Error(400, err)
	}

	if err := redis.AddQueue(newSub.ID); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}

	pubsub := redis.RDB.Subscribe(context.Background(), "Subscription")
	defer pubsub.Close()

	timeoutDuration := 10 * time.Second

	for {
		select {
		case msg := <-pubsub.Channel():
			parts := strings.Split(msg.Payload, " --- ")
			if len(parts) == 2 {
				subscriptionID, status := parts[0], parts[1]
				if subscriptionID == newSub.ID && status == "Success" {
					return c.Render(200, r.JSON(response))
				}
			}
		case <-time.After(timeoutDuration - time.Since(start)):
			return c.Render(200, r.JSON(map[string]string{"error": "Timeout"}))
		}
	}
}

func PostCancelHandler(c buffalo.Context) error {
	userid, err := middleware.GetCookieUserID(c)
	if err != nil {
		return c.Error(400, errors.New("no user in :"+err.Error()))
	}

	userPayment, err := redis.GetUserPayment(userid)
	if err != nil {
		return c.Error(400, err)
	} else if userPayment == nil {
		return c.Error(400, errors.New("no unarchived (active) subscriptions for user"))
	}

	stripeSub, err := sub.Get(userPayment.SubscriptionID, nil)
	if err != nil {
		return c.Error(400, err)
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	if _, err := sub.Update(stripeSub.ID, params); err != nil {
		return c.Error(400, err)
	}

	// Don't want to ACTUALLY error out for sending the email if everything else worked
	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
	if err == nil && firebaseUser.Email != "" {
		sendgrid.SendCancelEmail(firebaseUser.Email, true)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func PostUncancelHandler(c buffalo.Context) error {
	userid, err := middleware.GetCookieUserID(c)
	if err != nil {
		return c.Error(400, errors.New("no user in :"+err.Error()))
	}

	userPayment, err := redis.GetUserPayment(userid)
	if err != nil {
		return c.Error(400, err)
	} else if userPayment == nil {
		return c.Error(400, errors.New("no unarchived (active) subscriptions for user"))
	}

	stripeSub, err := sub.Get(userPayment.SubscriptionID, nil)
	if err != nil {
		return c.Error(400, err)
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	if _, err := sub.Update(stripeSub.ID, params); err != nil {
		return c.Error(400, err)
	}

	// Don't want to ACTUALLY error out for sending the email if everything else worked
	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
	if err == nil && firebaseUser.Email != "" {
		sendgrid.SendCancelEmail(firebaseUser.Email, false)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func PostUpdatePayment(c buffalo.Context) error {
	req := PaymentRequest{}
	if err := c.Bind(&req); err != nil {
		return c.Error(400, err)
	}

	userid, err := middleware.GetCookieUserID(c)
	if err != nil {
		return c.Error(400, errors.New("no user in :"+err.Error()))
	}

	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
	if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
		return c.Error(400, err)
	}

	userPayment, err := redis.GetUserPayment(userid)
	if err != nil {
		return c.Error(400, err)
	} else if userPayment == nil {
		return c.Error(400, errors.New("no unarchived (active) subscriptions for user"))
	}

	s, err := sub.Get(userPayment.SubscriptionID, nil)
	if err != nil {
		return c.Error(400, err)
	}

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(s.Customer.ID),
	}
	if _, err = paymentmethod.Attach(req.PaymentMethodID, params); err != nil {
		return c.Error(400, err)
	}

	customerParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
		},
	}
	if _, err = customer.Update(s.Customer.ID, customerParams); err != nil {
		return c.Error(400, err)
	}

	if _, err = sub.Update(s.ID, &stripe.SubscriptionParams{
		DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
	}); err != nil {
		return c.Error(400, err)
	}

	// Don't want to ACTUALLY error out for sending the email if everything else worked
	sendgrid.SendPaymentUpdateEmail(firebaseUser.Email)

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))

}
