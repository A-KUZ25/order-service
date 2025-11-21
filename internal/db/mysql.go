package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL() (*sql.DB, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("mysql open error: %w", err)
	}

	db.SetMaxOpenConns(50)                  // общее число открытых соединений
	db.SetMaxIdleConns(25)                  // простаивающих соединений
	db.SetConnMaxLifetime(time.Hour)        //сколько максимум живёт одно соединение, даже если активно
	db.SetConnMaxIdleTime(30 * time.Minute) //сколько максимум бездействует соединение

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping error: %w", err)
	}

	return db, nil
}
