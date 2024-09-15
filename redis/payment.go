package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stripe/stripe-go/v72/sub"
)

type UserPayment struct {
	CustomerID     string    `json:"c"`
	SubscriptionID string    `json:"s"`
	LastDate       time.Time `json:"d"`
	Active         bool      `json:"a"`
}

func GetUserPayment(uid string) (*UserPayment, error) {
	key := ":p:" + uid
	data, err := RDB.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var p UserPayment
	err = json.Unmarshal([]byte(data), &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func setUserPayment(uid string, p *UserPayment) error {
	key := ":p:" + uid
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = RDB.Set(context.Background(), key, data, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func setPaymentForUser(uid, subID string) error {
	return RDB.Set(context.Background(), ":b:"+subID, uid, 0).Err()
}

func GetUserBySubID(subID string) (string, error) {
	return RDB.Get(context.Background(), ":b:"+subID).Result()
}

func CheckUserPaying(uid string) (bool, error) {
	p, err := GetUserPayment(uid)
	if err != nil {
		return false, err
	}

	if p == nil {
		return false, nil
	}

	if !p.Active || p.CustomerID == "" || p.SubscriptionID == "" {
		return false, nil
	}

	if p.LastDate.After(time.Now()) {
		return true, nil
	}

	s, err := sub.Get(p.SubscriptionID, nil)
	if err != nil {
		return false, err
	}

	if s == nil {
		return false, errors.New("no actual active subscription")
	}

	periodEnd := time.Unix(s.CurrentPeriodEnd, 0)
	if s.Status == "active" && periodEnd.After(time.Now()) {
		p.LastDate = periodEnd
		if err := setUserPayment(uid, p); err != nil {
			return true, err
		}
		return true, nil
	}

	p.Active = false
	if err := setUserPayment(uid, p); err != nil {
		return false, err
	}

	return false, nil
}

func CreateBlankUserPayment(uid, custID string) error {
	p := UserPayment{CustomerID: custID}
	return setUserPayment(uid, &p)
}

func CreateSetUserPayment(uid, custID, subID string) error {
	p := UserPayment{CustomerID: custID, SubscriptionID: subID}
	if err := setUserPayment(uid, &p); err != nil {
		return err
	}
	return setPaymentForUser(uid, subID)
}

func SetSubOnUserPayment(uid, subID string) error {

	p, err := GetUserPayment(uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	p.SubscriptionID = subID
	p.Active = false

	if err := setUserPayment(uid, p); err != nil {
		return err
	}
	return setPaymentForUser(uid, subID)
}

func SetUserPaymentActive(uid, subID string, periodEnd time.Time) error {

	if periodEnd.Before(time.Now()) {
		return errors.New("period end date is before right now")
	}

	p, err := GetUserPayment(uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	if p.SubscriptionID != subID {
		return errors.New("incorrect subscription id for user")
	}

	p.Active = true
	p.LastDate = periodEnd

	return setUserPayment(uid, p)
}

func SetUserPaymentInactive(uid, subID string) error {

	p, err := GetUserPayment(uid)
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no user payment for this user")
	}

	if p.SubscriptionID != subID {
		return errors.New("incorrect subscription id for user")
	}

	p.Active = false

	return setUserPayment(uid, p)
}
