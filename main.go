package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Product struct {
	ID            string  `json:"product_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stock_quantity"`
}

var (
	redisClient *redis.Client
	wg          sync.WaitGroup
)

func main() {
	// create a redis connection

	redisClient = redis.NewClient(&redis.Options{})
	// Ping Redis to check connection
	if err := pingRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %s", err)
	}

	// Initialize routes

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// rest methods here:
	router.Get("/products/{id}", getProduct)
	router.Post("/products", createProduct)
	router.Put("/products/{id}", updateProduct)
	router.Delete("/products/{id}", deleteProduct)

	// Start HTTP server in a goroutine

	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	go func() {
		fmt.Println("Server starting on port 3000...")
		if err := server.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			log.Fatalf("Error starting server: %v", err)
		}

	}()

	// Handle graceful shutdowns

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs // Wait for termination signal

	// Shutdown the server gracefully

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Notify the WaitGroup to wait for all ongoing requests to finish
	wg.Wait()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

}

func pingRedis() error {
	_, err := redisClient.Ping(ctx).Result()

	if err != nil {
		return fmt.Errorf("failed to ping Redis (start redis connection): %w", err)
	}

	return nil
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// CRUD Handlers

func getProduct(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()
	productId := chi.URLParam(r, "id")
	val, err := redisClient.Get(ctx, productId).Result()
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	var product Product
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product)

}

func createProduct(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()
	var product Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(product)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := redisClient.Set(ctx, product.ID, data, 0).Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()
	productID := chi.URLParam(r, "id")
	var product Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// check if the product exists
	_, err := redisClient.Get(ctx, productID).Result()
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(product)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := redisClient.Set(ctx, productID, data, 0).Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	// check if the product exists
	_, err := redisClient.Get(ctx, productID).Result()
	if err != nil {
		http.Error(w, "You cannot delete product that doesnt exist", http.StatusNotFound)
		return
	}

	if err := redisClient.Del(ctx, productID).Err(); err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
