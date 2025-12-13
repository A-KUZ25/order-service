package main

import (
	"context"
	"log"
	"net/http"
	"orders-service/internal/db"
	"orders-service/internal/repository/mysql"
	"orders-service/internal/rest/orderhttp"
	"orders-service/order"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	//todo нет защиты от sql инъекций в фильтре
	mysqlDB, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("mysql init error: %v", err)
	}

	defer func() {
		if err := mysqlDB.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		} else {
			log.Println("db closed")
		}
	}()

	repo, err := mysql.NewOrdersRepository(mysqlDB)
	if err != nil {
		log.Fatalf("Repository init error: %v", err)
	}
	service := order.NewService(repo)
	handler := orderhttp.NewHandler(service)

	r := chi.NewRouter()
	orderhttp.RegisterRoutes(r, handler)

	port := os.Getenv("GO_PORT")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,  // время ожидания чтения запроса
		WriteTimeout: 10 * time.Second, // время на запись ответа
		IdleTimeout:  60 * time.Second, // время простоя соединения
	}

	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("server started at :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("shutdown signal received: %s", sig.String())
	case err := <-serverErrCh:
		if err != nil {
			log.Printf("server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// если graceful не удался, логируем и пробуем принудительно Close
		log.Printf("graceful shutdown failed: %v", err)
		if err := srv.Close(); err != nil {
			log.Printf("server close failed: %v", err)
		}
	} else {
		log.Println("server shutdown gracefully")
	}

}
