package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"order-service/internal/app"

	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := env("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5433/order_db?sslmode=disable")
	paymentGRPCAddr := env("PAYMENT_GRPC_ADDR", "localhost:50051")
	httpPort := env("PORT", "8081")
	grpcPort := env("GRPC_PORT", "50052")

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	deps := app.RouterDeps{
		DB: db,
		Config: app.Config{
			PaymentGRPCAddr: paymentGRPCAddr,
			GRPCPort:        grpcPort,
		},
	}

	grpcSrv := app.NewGRPCServer(db)
	go func() {
		if err := grpcSrv.Listen(fmt.Sprintf(":%s", grpcPort)); err != nil {
			log.Fatalf("order gRPC server error: %v", err)
		}
	}()

	router, cleanup, err := app.NewRouter(deps)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	log.Printf("order-service HTTP on :%s | gRPC streaming on :%s | payment-grpc=%s",
		httpPort, grpcPort, paymentGRPCAddr)
	if err := router.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
