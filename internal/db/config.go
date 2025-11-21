package db

import (
	"fmt"
	"os"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		Name:     os.Getenv("MYSQL_DB"),
	}

	if cfg.User == "" || cfg.Password == "" || cfg.Host == "" || cfg.Port == "" || cfg.Name == "" {
		return nil, fmt.Errorf("mysql config error: some environment variables are missing")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}
