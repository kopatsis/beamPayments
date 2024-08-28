package models

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func GetSubscription(uid string) (*Subscription, bool, error) {
	subscription := &Subscription{}

	err := DB.Where("user_id = ? AND archived = false", uid).First(subscription)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, true, nil
		}
		return nil, false, err
	}

	return subscription, false, nil
}

func CreateSubscription(userID, subscriptionID string, expiresDate time.Time) error {
	count, err := DB.Where("user_id = ? AND archived = false", userID).Count(&Subscription{})
	if err != nil {
		return err
	} else if count > 0 {
		return fmt.Errorf("an active subscription already exists for user_id: %s", userID)
	}

	sub := Subscription{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Processing:     true,
		Ending:         false,
		Paying:         false,
		EndDate:        time.Time{},
		ExpiresDate:    expiresDate,
		Archived:       false,
		ArchivedDate:   time.Time{},
	}

	return DB.Create(&sub)
}

func CancelSubscription(id int, endDate time.Time) error {
	return DB.RawQuery("UPDATE subscriptions SET ending = ?, end_date = ?, updated_at = ? WHERE id = ?", true, endDate, time.Now(), id).Exec()
}

func UncancelSubscription(id int) error {
	return DB.RawQuery("UPDATE subscriptions SET ending = ?, end_date = ?, updated_at = ? WHERE id = ?", false, time.Time{}, time.Now(), id).Exec()
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

func scheduledUpdate() {
	now := time.Now()
	endDateThreshold := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	expiresDateThreshold := now.Add(-72 * time.Hour).Format("2006-01-02 15:04:05")
	archivedDate := now.Format("2006-01-02 15:04:05")

	query := `
		UPDATE subscriptions 
		SET paying = false, archived = true, archived_date = ?
		WHERE (end_date <= ? OR expires_date <= ?)
	`

	if err := DB.RawQuery(query, archivedDate, endDateThreshold, expiresDateThreshold).Exec(); err != nil {
		log.Println("Error during scheduled update:", err)
	}
}

func ScheduledSubscriptionMods() {
	scheduledUpdate()
}
