package models

import (
	"log"
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

func ConfirmSubscirption(subscriptionID string, newExpiresDate time.Time) error {
	var sub Subscription

	err := DB.Where("subscription_id = ?", subscriptionID).First(&sub)
	if err != nil {
		return err
	}

	sub.Paying = true
	sub.Processing = false
	sub.ExpiresDate = newExpiresDate

	return DB.Save(&sub)
}

// func scheduledDelete() {
// 	now := time.Now()
// 	endDateThreshold := now.Add(-24 * time.Hour)
// 	expiresDateThreshold := now.AddDate(0, -1, 0)

// 	query := `
// 		DELETE FROM subscriptions
// 		WHERE (end_date <= ? OR expires_date <= ?)
// 	`

// 	if err := DB.RawQuery(query, endDateThreshold, expiresDateThreshold).Exec(); err != nil {
// 		log.Println("Error during scheduled delete:", err)
// 	}
// }

func scheduledUpdate() {
	now := time.Now()
	endDateThreshold := now.Add(-24 * time.Hour)
	expiresDateThreshold := now.Add(-72 * time.Hour)

	query := `
		UPDATE subscriptions 
		SET paying = false
		WHERE (end_date <= ? OR expires_date <= ?)
	`

	if err := DB.RawQuery(query, endDateThreshold, expiresDateThreshold).Exec(); err != nil {
		log.Println("Error during scheduled update:", err)
	}
}

func ScheduledSubscriptionMods() {
	scheduledUpdate()
	// scheduledDelete()
}
