package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"payment-service/internal/app"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := env("PAYMENT_DB_DSN", "postgres://postgres:postgres@localhost:5434/payment_db?sslmode=disable")
	httpPort := env("PORT", "8082")
	grpcPort := env("GRPC_PORT", "50051")

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	deps := app.RouterDeps{DB: db}

	grpcSrv := app.NewGRPCServer(deps)
	go func() {
		if err := grpcSrv.Listen(fmt.Sprintf(":%s", grpcPort)); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	router := app.NewRouter(deps)
	log.Printf("payment-service HTTP on :%s | gRPC on :%s", httpPort, grpcPort)
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
