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

func CancelSubscription(id int, endDate time.Time) error {
	return DB.RawQuery("UPDATE subscriptions SET ending = ?, end_date = ?, updated_at = ? WHERE id = ?", true, endDate, time.Now(), id).Exec()
}

func UncancelSubscription(id int) error {
	return DB.RawQuery("UPDATE subscriptions SET ending = ?, end_date = ?, updated_at = ? WHERE id = ?", false, time.Time{}, time.Now(), id).Exec()
}

func GetSubByUserID(userID string) (int, string, error) {
	var sub Subscription

	err := DB.Where("user_id = ?", userID).
		Select("id, subscription_id").
		First(&sub)

	if err != nil {
		return 0, "", err
	}

	return sub.ID, sub.SubscriptionID, nil
}
