package models

import "strings"

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
