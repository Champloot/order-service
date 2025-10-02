package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/config"
	"order-service/internal/cache"
	"order-service/internal/database"
	"order-service/internal/http"
	"order-service/internal/kafka"
)

func main() {
	// config
	cfg := config.LoadConfig()

	ctx := context.Background()

	// db init
	db, err := database.NewPostgresRepository(ctx, cfg.PostgresConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// cache init
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.CacheTTL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	defer redisCache.Close()

	// cache preload
	orders, err := db.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Failed to get orders for preloading cache: %v", err)
	} else {
		err = redisCache.PreloadOrders(ctx, orders)
		if err != nil {
			log.Printf("Failed to preload cache: %v", err)
		} else {
			log.Printf("Preloaded %d orders into cache", len(orders))
		}
	}

	// kafka consumer init
	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		db,
	)
	defer consumer.Close()

	go func() {
		// time for kafka load
		time.Sleep(10 * time.Second)
		consumer.Start(ctx)
	}()

	// http server init
	httpServer := http.NewServer(redisCache, db)
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.HTTPAddr)
		if err := httpServer.Start(cfg.HTTPAddr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Println("Service started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down service...")
}
