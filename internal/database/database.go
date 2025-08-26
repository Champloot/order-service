package database

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	// "time"

	// "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, connString string) (*PostgresRepository, error){
	// Parse connection string
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection string: %w", err)
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %w", err)
	}

	repo := &PostgresRepository{pool: pool}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return repo, nil
}

func (r *PostgresRepository) createTables(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			order_uid VARCHAR(255) PRIMARY KEY,
			track_number VARCHAR(255),
			entry VARCHAR(255),
			delivery JSONB NOT NULL,
			payment JSONB NOT NULL,
			items JSONB NOT NULL,
			locale VARCHAR(255),
			internal_signature VARCHAR(255),
			customer_id VARCHAR(255),
			delivery_service VARCHAR(255),
			shardkey VARCHAR(255),
			sm_id INTEGER,
			date_created TIMESTAMP WITH TIME ZONE,
			oof_shard VARCHAR(255)
		);

		CREATE INDEX IF NOT EXISTS idx_orders_order_uid ON orders(order_uid);
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("Failed to create tables: %w", err)
	}

	log.Println("Tables created or already exist")
	return nil
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}