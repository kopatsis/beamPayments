package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func AddEmail(email string) (string, error) {
	id := uuid.New().String()
	key := ":e:" + id
	err := RDB.Set(context.Background(), key, email, 0).Err()
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetAndDeleteEmail(uuid string) (string, error) {
	key := ":e:" + uuid
	email, err := RDB.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", errors.New("dne before deletion")
	}
	if err != nil {
		return "", err
	}
	err = RDB.Del(context.Background(), key).Err()
	if err == redis.Nil {
		return email, errors.New("dne after deletion")
	}
	if err != nil {
		return email, err
	}
	return email, nil
}
