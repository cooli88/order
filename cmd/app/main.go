package main

import (
	"log"
	"net/http"
	"os"

	"github.com/demo/contracts/gen/go/order/v1/orderv1connect"
	"github.com/demo/order/internal/domain/orders"
	"github.com/demo/order/internal/store"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"
	}

	pgStore, err := store.NewPostgresStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgStore.Close()

	orderService := orders.NewServer(pgStore)

	mux := http.NewServeMux()
	path, handler := orderv1connect.NewOrderServiceHandler(orderService)
	mux.Handle(path, handler)

	addr := ":8081"
	log.Printf("Order service listening on %s", addr)

	err = http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{}))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
