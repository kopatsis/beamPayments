package actions

import (
	"beam_payments/actions/firebaseApp"
	"beam_payments/models"
	"context"
	"errors"
	"os"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/invoice"
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

	userid := c.Session().Get("user_id").(string)
	if userid == "" {
		return c.Error(400, errors.New("no user in "))
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

	if err := models.CreateSubscription(userid, newSub.ID, time.Unix(newSub.CurrentPeriodEnd, 0)); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}
	elapsed := time.Duration(0)

	for elapsed < 6*time.Second {
		inv, err := invoice.Get(newSub.LatestInvoice.ID, nil)
		if err != nil {
			return c.Render(200, r.JSON(response))
		}

		if inv.Status == "paid" {
			return c.Render(200, r.JSON(response))
		}

		time.Sleep(500 * time.Millisecond)

		elapsed = time.Since(start)
	}

	return c.Render(200, r.JSON(response))
}

func PostCancelHandler(c buffalo.Context) error {
	userid := c.Session().Get("user_id").(string)
	if userid == "" {
		return c.Error(400, errors.New("no user in "))
	}

	databaseID, subID, err := models.GetSubByUserID(userid)
	if err != nil {
		return c.Error(400, err)
	}

	stripeSub, err := sub.Get(subID, nil)
	if err != nil {
		return c.Error(400, err)
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	if _, err := sub.Update(stripeSub.ID, params); err != nil {
		return c.Error(400, err)
	}

	if err := models.CancelSubscription(databaseID, time.Unix(stripeSub.CurrentPeriodEnd, 0)); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func PostUncancelHandler(c buffalo.Context) error {
	userid := c.Session().Get("user_id").(string)
	if userid == "" {
		return c.Error(400, errors.New("no user in "))
	}

	databaseID, subID, err := models.GetSubByUserID(userid)
	if err != nil {
		return c.Error(400, err)
	}

	stripeSub, err := sub.Get(subID, nil)
	if err != nil {
		return c.Error(400, err)
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	if _, err := sub.Update(stripeSub.ID, params); err != nil {
		return c.Error(400, err)
	}

	if err := models.UncancelSubscription(databaseID); err != nil {
		return c.Error(400, err)
	}

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func PostUpdatePayment(c buffalo.Context) error {
	req := PaymentRequest{}
	if err := c.Bind(&req); err != nil {
		return c.Error(400, err)
	}

	userid := c.Session().Get("user_id").(string)
	if userid == "" {
		return c.Error(400, errors.New("no user in "))
	}

	firebaseUser, err := firebaseApp.FirebaseAuth.GetUser(context.Background(), userid)
	if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
		return c.Error(400, err)
	}

	_, subID, err := models.GetSubByUserID(userid)
	if err != nil {
		return c.Error(400, err)
	}

	s, err := sub.Get(subID, nil)
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

	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))

}
