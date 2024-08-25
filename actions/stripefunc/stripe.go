package stripefunc

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

func GetPaymentMethodDetails(subscriptionID string) (string, string, string, error) {
	subscription, err := sub.Get(subscriptionID, nil)
	if err == nil && subscription.DefaultPaymentMethod != nil {
		paymentMethod, err := paymentmethod.Get(subscription.DefaultPaymentMethod.ID, nil)
		if err == nil {
			switch paymentMethod.Type {
			case stripe.PaymentMethodTypeCard:
				card := paymentMethod.Card
				return "Card", string(card.Brand), card.Last4, nil
			default:
				return string(paymentMethod.Type), "", "", nil
			}
		}
	}

	params := &stripe.InvoiceListParams{
		Subscription: stripe.String(subscriptionID),
	}
	params.Filters.AddFilter("limit", "", "10")

	i := invoice.List(params)

	for i.Next() {
		inv := i.Invoice()
		if inv.PaymentIntent != nil {
			paymentIntent, err := paymentintent.Get(inv.PaymentIntent.ID, nil)
			if err == nil && paymentIntent.PaymentMethod != nil {
				paymentMethodID := paymentIntent.PaymentMethod.ID
				if paymentMethodID != "" {
					paymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
					if err == nil {
						switch paymentMethod.Type {
						case stripe.PaymentMethodTypeCard:
							card := paymentMethod.Card
							return "Card", string(card.Brand), card.Last4, nil
						default:
							return string(paymentMethod.Type), "", "", nil
						}
					}
				}
			}
		}
	}

	return "", "", "", fmt.Errorf("no valid payment method found for subscription ID: %s", subscriptionID)
}
