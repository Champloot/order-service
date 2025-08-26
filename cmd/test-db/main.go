package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Конфигурация подключения к базе данных
	connString := "postgres://user:password@localhost:5432/orderservice?sslmode=disable"
	
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Парсим строку подключения
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Failed to parse connection string: %v", err)
	}

	// Создаем пул подключений
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	// Создаем таблицу
	if err := createTables(ctx, pool); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	fmt.Println("Tables created successfully!")

	// Проверяем существование таблицы
	exists, err := checkTableExists(ctx, pool, "orders")
	if err != nil {
		log.Fatalf("Failed to check table existence: %v", err)
	}

	if exists {
		fmt.Println("Table 'orders' exists in the database")
	} else {
		fmt.Println("Table 'orders' does NOT exist in the database")
	}

	fmt.Println("Connection and table creation test completed successfully!")
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			order_uid VARCHAR(255) PRIMARY KEY,
			track_number VARCHAR(255),
			entry VARCHAR(255),
			delivery JSONB NOT NULL,
			payment JSONB NOT NULL,
			items JSONB NOT NULL,
			locale VARCHAR(10),
			internal_signature VARCHAR(255),
			customer_id VARCHAR(255),
			delivery_service VARCHAR(255),
			shardkey VARCHAR(255),
			sm_id INTEGER,
			date_created TIMESTAMP WITH TIME ZONE,
			oof_shard VARCHAR(255)
		)
	`

	_, err := pool.Exec(ctx, query)
	return err
}

func checkTableExists(ctx context.Context, pool *pgxpool.Pool, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = $1
		)
	`

	var exists bool
	err := pool.QueryRow(ctx, query, tableName).Scan(&exists)
	return exists, err
}
