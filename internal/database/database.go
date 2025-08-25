package database

import (
	"context"
	"fmt"
	"log"
	"time"

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

	repo := $PostgresRepository{pool: pool}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return repo, nil
}