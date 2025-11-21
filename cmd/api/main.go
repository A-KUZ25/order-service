package main

import (
	"log"
	"net/http"
	"orders-service/internal/db"
	"orders-service/internal/repository/mysql"
	"orders-service/internal/rest/orderhttp"
	"orders-service/order"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	mysqlDB, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("mysql init error: %v", err)
	}

	repo, err := mysql.NewOrdersRepository(mysqlDB)
	if err != nil {
		log.Fatalf("Repository init error: %v", err)
	}
	service := order.NewService(repo)
	handler := orderhttp.NewHandler(service)

	r := chi.NewRouter()
	orderhttp.RegisterRoutes(r, handler)
	port := os.Getenv("GO_PORT")
	log.Printf("server started at :%s\n", port)
	http.ListenAndServe(":"+port, r)

	_ = mysqlDB
}
