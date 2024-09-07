package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func AddQueue(sub string) error {
	key := ":s:" + sub
	value := []byte{1}
	err := RDB.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetQueue(sub string) (bool, error) {
	key := ":s:" + sub
	value, err := RDB.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if value == string([]byte{1}) {
		return true, nil
	}
	return false, nil
}

func DeleteQueue(sub string) error {
	key := ":s:" + sub
	err := RDB.Del(context.Background(), key).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}
