package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/demo/order/internal/entity"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderStore interface {
	Create(ctx context.Context, order *entity.Order) error
	Get(ctx context.Context, id string) (*entity.Order, error)
	List(ctx context.Context) ([]*entity.Order, error)
	Close() error
}

type PostgresStore struct {
	db *sqlx.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func createSchema(db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			item VARCHAR(255) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`
	_, err := db.Exec(query)
	return err
}

func (s *PostgresStore) Create(ctx context.Context, order *entity.Order) error {
	const query = `
		INSERT INTO orders (id, user_id, item, amount, status, created_at)
		VALUES (:id, :user_id, :item, :amount, :status, :created_at)`
	_, err := s.db.NamedExecContext(ctx, query, order)
	return err
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*entity.Order, error) {
	const query = `SELECT id, user_id, item, amount, status, created_at FROM orders WHERE id = $1`
	var order entity.Order
	err := s.db.GetContext(ctx, &order, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *PostgresStore) List(ctx context.Context) ([]*entity.Order, error) {
	const query = `SELECT id, user_id, item, amount, status, created_at FROM orders ORDER BY created_at DESC`
	var orders []*entity.Order
	err := s.db.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
