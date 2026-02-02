package store

import (
	"context"

	"github.com/demo/order/internal/entity"
)

// MockOrderStore is a mock implementation of OrderStore for testing.
type MockOrderStore struct {
	CreateFunc func(ctx context.Context, order *entity.Order) error
	GetFunc    func(ctx context.Context, id string) (*entity.Order, error)
	ListFunc   func(ctx context.Context) ([]*entity.Order, error)
	CloseFunc  func() error
}

func (m *MockOrderStore) Create(ctx context.Context, order *entity.Order) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order)
	}
	return nil
}

func (m *MockOrderStore) Get(ctx context.Context, id string) (*entity.Order, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockOrderStore) List(ctx context.Context) ([]*entity.Order, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockOrderStore) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
