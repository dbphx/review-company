package db

import (
	"context"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/review-company/backend/internal/config"
)

var RDB *redis.Client

func ConnectRedis(cfg *config.Config) {
	dbIndex, err := strconv.Atoi(cfg.RedisDB)
	if err != nil {
		dbIndex = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       dbIndex,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect Redis: %v", err)
	}

	RDB = client
	log.Println("Connected to Redis successfully")
}
