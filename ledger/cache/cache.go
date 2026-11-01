package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

func NewCache() (*Cache, error) {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	db := getIntEnv("REDIS_DB", 0)
	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})

	ctx := context.Background()

	// Проверка подключения
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s (DB: %d)", addr, db)

	return &Cache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (c *Cache) Get(key string, value interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}

	if err := json.Unmarshal(data, value); err != nil {
		return err
	}

	log.Printf("Cache hit: %s", key)
	return nil
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := c.client.Set(c.ctx, key, data, ttl).Err(); err != nil {
		return err
	}

	log.Printf("Cache set: %s (TTL: %v)", key, ttl)
	return nil
}

func (c *Cache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

func (c *Cache) Close() error {
	log.Println("Closing Redis connection")
	return c.client.Close()
}

var ErrCacheMiss = fmt.Errorf("cache miss")

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
