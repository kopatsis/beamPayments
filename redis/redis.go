package redis

import (
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func init() {
	redisAddr, redisPass := os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD")
	if redisAddr == "" || redisPass == "" {
		log.Fatalf("cannot connect to redis as no addr and/or pass present in env")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: "default",
		Password: redisPass,
		DB:       0,
	})

	RDB = rdb
}
