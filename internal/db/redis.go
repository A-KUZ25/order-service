package db

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host              string
	Port              string
	Database          int
	Password          string
	ConnectionTimeout time.Duration
	DataTimeout       time.Duration
}

func NewRedisActiveOrders() (*redis.Client, error) {
	cfg, err := LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DB:           cfg.Database,
		Password:     cfg.Password,
		DialTimeout:  cfg.ConnectionTimeout,
		ReadTimeout:  cfg.DataTimeout,
		WriteTimeout: cfg.DataTimeout,
	})

	return client, nil
}

func LoadRedisConfig() (*RedisConfig, error) {
	host := os.Getenv("REDIS_MAIN_HOST")
	port := os.Getenv("REDIS_MAIN_PORT")
	dbRaw := os.Getenv("REDIS_MAIN_DATABASE_ORDERS_ACTIVE")

	if host == "" || port == "" || dbRaw == "" {
		return nil, fmt.Errorf("redis config error: some environment variables are missing")
	}

	database, err := strconv.Atoi(dbRaw)
	if err != nil {
		return nil, fmt.Errorf("redis config error: invalid database: %w", err)
	}

	connectionTimeout, err := parseRedisTimeout("REDIS_MAIN_CONNECT_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	dataTimeout, err := parseRedisTimeout("REDIS_MAIN_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &RedisConfig{
		Host:              host,
		Port:              port,
		Database:          database,
		Password:          os.Getenv("REDIS_MAIN_PASSWORD"),
		ConnectionTimeout: connectionTimeout,
		DataTimeout:       dataTimeout,
	}, nil
}

func parseRedisTimeout(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("redis config error: invalid %s: %w", name, err)
	}

	return time.Duration(seconds) * time.Second, nil
}
