package database

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"log"
	"time"

	"order-service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrOrderNotFound = errors.New("Order not found")
)

type Querier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

// Wrapper for transactions with methods Commit/Rollback
type TxWrapper struct {
	tx pgx.Tx
}

type DatabaseConfig struct {
	URL					string
	MaxConns          	int
	MinConns          	int
	MaxConnLifetime   	time.Duration
	MaxConnIdleTime   	time.Duration
	HealthCheckPeriod 	time.Duration
}

func NewPostgresRepository(ctx context.Context, config DatabaseConfig) (*PostgresRepository, error) {
	// Parse connection string
	config, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection string: %w", err)
	}

	// set params for pool
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

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

	// Create tables
	if err := repo.createTables(ctx); err != nil {
		return nil, fmt.Errorf("Failed to init schema: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return repo, nil
}

func (r *PostgresRepository) BeginTx(ctx context.Context) (*TxWrapper, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:	pgx.ReadCommitted,
		AccessMode:	pgx.ReadWrite,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to begin transaction: %w", err)
	}
	return &TxWrapper{tx: tx}, nil
}

// commit of transaction
func (tw *TxWrapper) Commit(ctx context.Context) error {
	if err := tw.tx.Commit(ctx); err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

// rollback of transaction
func (tw *TxWrapper) Rollback(ctx context.Context) error {
	if err := tw.tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("Failed to rollback transaction: %w", err)
	}
	return nil
}

func (tw *TxWrapper) GetQuerier() Querier {
	return tw.tx
}

func (r *PostgresRepository) WithTransaction(ctx context.Context, fn func(tx *TxWrapper) error) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("Transaction error: %w, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return nil
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
		CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
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
	return r.saveOrder(ctx, r.pool, order)
}

// with transaction
func (r *PostgresRepository) SaveOrderTx(ctx context.Context, tx *TxWrapper, order *models.Order) error {
	return r.saveOrder(ctx, tx.GetQuerier(), order)
}

func (r *PostgresRepository) saveOrder(ctx context.Context, querier Querier, order *models.Order) error {
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
		return fmt.Errorf("Failed to marshal delivery: %w", err)
	}

	paymentJSON, err := json.Marshal(order.Payment)
	if err != nil {
		return fmt.Errorf("Failed to marshal payment: %w", err)
	}

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("Failed to marshal items: %w", err)
	}

	_, err = querier.Exec(
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
		return fmt.Errorf("Failed to save order: %w", err)
	}

	log.Printf("Order %s saved successfully", order.OrderUID)
	return nil
}

func (r *PostgresRepository) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	return r.getOrder(ctx, r.pool, orderUID)
}

// transaction version
func (r *PostgresRepository) GetOrderTx(ctx context.Context, tx *TxWrapper, orderUID string) (*models.Order, error) {
	return r.getOrder(ctx, tx.GetQuerier(), orderUID)
}

func (r *PostgresRepository) getOrder(ctx context.Context, querier Querier, orderUID string) (*models.Order, error) {
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

	err := querier.QueryRow(ctx, query, orderUID).Scan(
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

	// use error.is
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrderNotFound
	} else if err != nil {
		return nil, fmt.Errorf("Failed to get order: %w", err)
	}

	order.DateCreated = DateCreated

	// Parse JSON fields
	if err := json.Unmarshal(deliveryJSON, &order.Delivery); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal delivery: %w", err)
	}

	if err := json.Unmarshal(paymentJSON, &order.Payment); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal payment: %w", err)
	}

	if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal items: %w", err)
	}

	return &order, nil
}

func (r *PostgresRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	return r.getAllOrders(ctx, r.pool)
}

// transaction version
func (r *PostgresRepository) GetAllOrdersTx(ctx context.Context, tx *TxWrapper) ([]models.Order, error) {
	return r.getAllOrders(ctx, tx.GetQuerier())
}

func (r *PostgresRepository) getAllOrders(ctx context.Context, querier Querier) ([]models.Order, error) {
	query := `
		SELECT
			order_uid, track_number, entry, delivery, payment, items,
			locale, internal_signature, customer_id, delivery_service,
			shardkey, sm_id, date_created, oof_shard
		FROM orders
		ORDER BY date_created DESC
	`

	rows, err := querier.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		var deliveryJSON, paymentJSON, itemsJSON []byte
		var dateCreated time.Time

		err := rows.Scan(
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
			&dateCreated,
			&order.OofShard,
		)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan order: %w", err)
		}

		order.DateCreated = dateCreated

		// Parse JSON fields
		if err := json.Unmarshal(deliveryJSON, &order.Delivery); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal delivery: %w", err)
		}

		if err := json.Unmarshal(paymentJSON, &order.Payment); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal payment: %w", err)
		}

		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal items: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating orders: %w", err)
	}

	return orders, nil
}

func (r *PostgresRepository) DeleteOrder(ctx context.Context, orderUID string) error {
	return r.deleteOrder(ctx, r.pool, orderUID)
}

func (r *PostgresRepository) DeleteOrderTx(ctx context.Context, tx *TxWrapper, orderUID string) error {
	return r.deleteOrder(ctx, tx.GetQuerier(), orderUID)
}

func (r *PostgresRepository) deleteOrder(ctx context.Context, querier Querier, orderUID string) error {
	query := `DELETE FROM orders WHERE order_uid = $1`

	result, err := querier.Exec(ctx, query, orderUID)
	if err != nil {
		return fmt.Errorf("Failed to delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrOrderNotFound
	}

	log.Printf("Order %s deleted successfully", orderUID)
	return nil
}
