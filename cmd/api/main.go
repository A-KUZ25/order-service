package main

import (
	"context"
	"net/http"
	"orders-service/internal/app/order"
	"orders-service/internal/app/orderformat"
	"orders-service/internal/app/orderview"
	"orders-service/internal/db"
	"orders-service/internal/legacy/address"
	"orders-service/internal/logging"
	"orders-service/internal/repository/mysql"
	"orders-service/internal/repository/redisactive"
	"orders-service/internal/rest/orderhttp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	logging.Init("orders-service")

	//todo нет защиты от sql инъекций в фильтре
	mysqlDB, err := db.NewMySQL()
	if err != nil {
		logging.Error(context.Background(), "mysql init error", err)
		os.Exit(1)
	}

	defer func() {
		if err := mysqlDB.Close(); err != nil {
			logging.Error(context.Background(), "error closing db", err)
		} else {
			logging.Info(context.Background(), "db closed")
		}
	}()

	redisClient, err := db.NewRedisActiveOrders()
	if err != nil {
		logging.Error(context.Background(), "redis init error", err)
		os.Exit(1)
	}

	defer func() {
		if err := redisClient.Close(); err != nil {
			logging.Error(context.Background(), "error closing redis", err)
		} else {
			logging.Info(context.Background(), "redis closed")
		}
	}()

	repo, err := mysql.NewOrdersRepository(mysqlDB)
	if err != nil {
		logging.Error(context.Background(), "repository init error", err)
		os.Exit(1)
	}
	service := order.NewService(
		repo,
		redisactive.NewActiveOrdersRepository(redisClient),
		orderformat.NewAddressResolver(address.NewParser()),
		orderview.NewAssembler(
			redisactive.NewActiveOrdersRepository(redisClient),
			mysql.NewStatusTranslator(mysqlDB),
			mysql.NewShowOrderCodeProvider(mysqlDB),
		),
	)
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
		logging.Info(context.Background(), "server started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logging.Info(context.Background(), "shutdown signal received", "signal", sig.String())
	case err := <-serverErrCh:
		if err != nil {
			logging.Error(context.Background(), "server error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Error(context.Background(), "graceful shutdown failed", err)
		if err := srv.Close(); err != nil {
			logging.Error(context.Background(), "server close failed", err)
		}
	} else {
		logging.Info(context.Background(), "server shutdown gracefully")
	}

}
