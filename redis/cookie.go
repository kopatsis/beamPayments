package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type CookieLimit struct {
	Success   bool      `json:"s"`
	Banned    bool      `json:"b"`
	ResetDate time.Time `json:"r"`
}

func getCookieLimit(uid string) (CookieLimit, error) {
	key := ":c:" + uid
	data, err := RDB.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return CookieLimit{}, nil
	}
	if err != nil {
		return CookieLimit{}, err
	}

	var limit CookieLimit
	err = json.Unmarshal([]byte(data), &limit)
	if err != nil {
		return CookieLimit{}, err
	}

	if !limit.Success {
		return CookieLimit{}, errors.New("not unmarshalled correclty")
	}

	return limit, nil
}

func addCookieLimit(uid string, limit CookieLimit) error {
	key := ":c:" + uid
	data, err := json.Marshal(limit)
	if err != nil {
		return err
	}

	err = RDB.Set(context.Background(), key, data, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func AddResetDate(uid string) error {
	cl, err := getCookieLimit(uid)
	if err != nil {
		return err
	}

	cl.Success = true
	cl.ResetDate = time.Now()

	return addCookieLimit(uid, cl)
}

func AddBanned(uid string) error {
	cl, err := getCookieLimit(uid)
	if err != nil {
		return err
	}

	cl.Success = true
	cl.Banned = true

	return addCookieLimit(uid, cl)
}

func CheckCookeLimit(uid string, added time.Time) (banned bool, reset bool, reterr error) {

	cl, err := getCookieLimit(uid)
	if err != nil {
		return false, false, err
	}

	if !cl.Success {
		return false, false, nil
	}

	if cl.Banned {
		return true, false, nil
	}

	if added.Before(cl.ResetDate) {
		return false, true, nil
	}

	return false, false, nil
}
