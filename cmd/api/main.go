package main

import (
	"log"
	"net/http"
	"orders-service/internal/db"
	"orders-service/internal/rest/orders"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	mysqlDB, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("mysql init error: %v", err)
	}

	// DI: создаём handler (пока без service, заглушка)
	handler := &orders.Handler{
		// service: добавим позже
	}

	r := chi.NewRouter()
	orders.RegisterRoutes(r, handler)

	log.Println("server started at :8080")
	http.ListenAndServe(":8080", r)

	_ = mysqlDB
}
