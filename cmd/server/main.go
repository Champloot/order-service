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
	"order-service/internal/ports"
)

func main() {
	// config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting service in %s mode", cfg.App.Env)

	ctx := context.Background()

	// db init
	var orderRepository ports.OrderRepository
	orderRepository, err = database.NewPostgresRepository(ctx, database.DatabaseConfig{
		URL:               cfg.Database.URL,
		MaxConns:          cfg.Database.MaxConns,
		MinConns:          cfg.Database.MinConns,
		MaxConnLifetime:   cfg.Database.MaxConnLifetime,
		MaxConnIdleTime:   cfg.Database.MaxConnIdleTime,
		HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer orderRepository.(*database.PostgresRepository).Close()

	// cache init
	var orderCache ports.OrderCache
	orderCache, err = cache.NewRedisCache(
		cfg.Cache.Addr,
		cfg.Cache.Password,
		cfg.Cache.DB,
		cfg.Cache.TTL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	defer orderCache.Close()

	// cache preload
	orders, err := orderRepository.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Failed to get orders for preloading cache: %v", err)
	} else {
		err = orderCache.PreloadOrders(ctx, orders)
		if err != nil {
			log.Printf("Failed to preload cache: %v", err)
		} else {
			log.Printf("Preloaded %d orders into cache", len(orders))
		}
	}

	// kafka consumer init
	var orderConsumer ports.OrderConsumer
	orderConsumer = kafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		orderRepository,
		cfg.Consumer.Timeout,
		cfg.Consumer.MinBytes,
		cfg.Consumer.MaxBytes,
		cfg.Consumer.MaxWait,
		cfg.Consumer.RetryDelay,
	)
	defer orderConsumer.Close()

	go func() {
		// time for kafka load
		time.Sleep(10 * time.Second)
		orderConsumer.Start(ctx)
	}()

	// http server init
	var httpServer ports.HTTPServer
	httpServer = http.NewServer(orderCache, orderRepository)
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.HTTP.Addr)
		if err := httpServer.Start(cfg.HTTP.Addr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Println("Service started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down service...")
}
