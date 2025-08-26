package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"order-service/internal/models"
	"github.com/jackc/pgx/v5"
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
				CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created);
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

func (r *PostgresRepository) SaveOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (
			order_uid, track_number, entry, delivery, payment, items,
			locale, internal_signature, customer_id, delivery_service,
			shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			delivery = EXCLUDED.delivery,
			payment = EXCLUDED.payment,
			items = EXCLUDED.items,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard
	`

	// Converted to JSON
	deliveryJSON, err := json.Marshal(order.Delivery)
	if err != nil {
		return fmt.Errorf("failed to marshal delivery: %w", err)
	}

	paymentJSON, err := json.Marshal(order.Payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	_, err = r.pool.Exec(
		ctx,
		query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		deliveryJSON,
		paymentJSON,
		itemsJSON,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	log.Printf("Order %s saved successfully", order.OrderUID)
	return nil
}

func (r *PostgresRepository) GetOrder(ctx context.Context, OrderUID string) (*models.Order, error) {
	query := `
		SELECT
			order_uid, track_number, entry, delivery, payment, items,
			locale, internal_signature, customer_id, delivery_service,
			shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`

	var order models.Order
	var deliveryJSON, paymentJSON, itemsJSON []byte
	var DateCreated time.Time

	err := r.pool.QueryRow(ctx, query, OrderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&deliveryJSON,
		&paymentJSON,
		&itemsJSON,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&DateCreated,
		&order.OofShard,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.DateCreated = DateCreated

	// Parse JSON fields
	if err := json.Unmarshal(deliveryJSON, &order.Delivery); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delivery: %w", err)
	}

	if err := json.Unmarshal(paymentJSON, &order.Payment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment: %w", err)
	}

	if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return &order, nil
}