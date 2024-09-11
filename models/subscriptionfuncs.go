package models

import (
	"beam_payments/actions/sendgrid"
	"beam_payments/redis"
	"fmt"
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

func GetSubscriptionBySubID(subID string) (*Subscription, bool, error) {
	subscription := &Subscription{}

	err := DB.Where("subscription_id = ?", subID).First(subscription)
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

func ConfirmSubscription(subscriptionID string, newExpiresDate time.Time) (string, error) {
	var sub Subscription

	err := DB.Where("subscription_id = ?", subscriptionID).First(&sub)
	if err != nil {
		return "", err
	}

	sub.Paying = true
	sub.Processing = false
	sub.ExpiresDate = newExpiresDate

	err = DB.Save(&sub)
	if err != nil {
		return "", err
	}

	return sub.UserID, nil
}

func scheduledUpdate(refTime time.Time) error {
	endDateThreshold := refTime.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	expiresDateThreshold := refTime.Add(-72 * time.Hour).Format("2006-01-02 15:04:05")
	archivedDate := time.Now().Format("2006-01-02 15:04:05")

	query := `
		UPDATE subscriptions 
		SET paying = false, archived = true, archived_date = ?
		WHERE (end_date <= ? OR expires_date <= ?)
	`

	if err := DB.RawQuery(query, archivedDate, endDateThreshold, expiresDateThreshold).Exec(); err != nil {
		return err
	}

	return nil
}

func scheduledGetUserIDs() (userIDs []string, now time.Time, err error) {
	now = time.Now()
	endDateThreshold := now.Add(-24 * time.Hour)
	expiresDateThreshold := now.Add(-72 * time.Hour)

	type UserIDOnly struct {
		UserID string `db:"user_id"`
	}

	var results []UserIDOnly

	err = DB.RawQuery(`
		SELECT user_id 
		FROM subscriptions 
		WHERE end_date <= ? OR expires_date <= ?
	`, endDateThreshold, expiresDateThreshold).All(&results)

	if err != nil {
		return nil, now, err
	}

	for _, result := range results {
		userIDs = append(userIDs, ":u:"+result.UserID)
	}

	return userIDs, now, nil
}

func ScheduledSubscriptionMods() {
	ids, ref, err := scheduledGetUserIDs()
	if err != nil {
		err = sendgrid.SendSeriousErrorAlert("Scheduled Upating Get IDs", "This error: "+err.Error())
		if err != nil {
			sendgrid.SendSeriousErrorAlert("Sending the Actual Issue Email", "This error: "+err.Error())
		}
	}
	if len(ids) > 0 {
		err := scheduledUpdate(ref)
		if err != nil {
			err = sendgrid.SendSeriousErrorAlert("Scheduled Upating IDs", "This error: "+err.Error())
			if err != nil {
				sendgrid.SendSeriousErrorAlert("Sending the Actual Issue Email", "This error: "+err.Error())
			}
		}

		err = redis.DeleteSubMass(ids)
		if err != nil {
			err = sendgrid.SendSeriousErrorAlert("Scheduled Upating Users on Redis", "This error: "+err.Error())
			if err != nil {
				sendgrid.SendSeriousErrorAlert("Sending the Actual Issue Email", "This error: "+err.Error())
			}
		}
	}

}
