package models

import (
	"strings"
	"time"
)

func GetSubscription(uid string) (*Subscription, bool, error) {
	subscription := &Subscription{}

	err := DB.Where("user_id = ?", uid).First(subscription)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, true, nil
		}
		return nil, false, err
	}

	return subscription, false, nil
}

func CreateSubscription(userID, subscriptionID string, expiresDate time.Time) error {
	sub := Subscription{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Processing:     true,
		Ending:         false,
		Paying:         false,
		EndDate:        time.Time{},
		ExpiresDate:    expiresDate,
	}

	return DB.Create(&sub)
}
