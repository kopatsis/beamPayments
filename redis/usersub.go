package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func AddSub(sub string) error {
	key := ":u:" + sub
	value := []byte{1}
	err := RDB.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetSub(sub string) (bool, error) {
	key := ":u:" + sub
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

func DeleteSubMass(keys []string) error {
	err := RDB.Del(context.Background(), keys...).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}
