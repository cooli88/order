package orders

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/entity"
	"github.com/demo/order/internal/store"
	"github.com/stretchr/testify/require"
)

func TestCheckOrderOwnerHandler(t *testing.T) {
	tests := []struct {
		name         string
		request      *orderv1.CheckOrderOwnerRequest
		setupMock    func() *store.MockOrderStore
		wantErr      bool
		wantCode     connect.Code
		validateResp func(t *testing.T, resp *connect.Response[orderv1.CheckOrderOwnerResponse])
	}{
		{
			name: "success",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "order-123",
				UserId:  "user-456",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					GetFunc: func(ctx context.Context, id string) (*entity.Order, error) {
						require.Equal(t, "order-123", id)
						return &entity.Order{
							ID:        "order-123",
							UserID:    "user-456",
							Item:      "Test Item",
							Amount:    49.99,
							Status:    entity.OrderStatusNew,
							CreatedAt: time.Now().UTC(),
						}, nil
					},
				}
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.CheckOrderOwnerResponse]) {
				require.NotNil(t, resp)
			},
		},
		{
			name: "empty_order_id",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "",
				UserId:  "user-456",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "empty_user_id",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "order-123",
				UserId:  "",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "not_found",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "non-existent-order",
				UserId:  "user-456",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					GetFunc: func(ctx context.Context, id string) (*entity.Order, error) {
						return nil, store.ErrOrderNotFound
					},
				}
			},
			wantErr:  true,
			wantCode: connect.CodeNotFound,
		},
		{
			name: "wrong_owner",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "order-123",
				UserId:  "different-user",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					GetFunc: func(ctx context.Context, id string) (*entity.Order, error) {
						return &entity.Order{
							ID:        "order-123",
							UserID:    "original-user",
							Item:      "Test Item",
							Amount:    49.99,
							Status:    entity.OrderStatusNew,
							CreatedAt: time.Now().UTC(),
						}, nil
					},
				}
			},
			wantErr:  true,
			wantCode: connect.CodePermissionDenied,
		},
		{
			name: "store_error",
			request: &orderv1.CheckOrderOwnerRequest{
				OrderId: "order-123",
				UserId:  "user-456",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					GetFunc: func(ctx context.Context, id string) (*entity.Order, error) {
						return nil, errors.New("database connection failed")
					},
				}
			},
			wantErr:  true,
			wantCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := tt.setupMock()
			handler := newCheckOrderOwnerHandler(mockStore)

			req := connect.NewRequest(tt.request)
			resp, err := handler.Handle(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, tt.wantCode, connectErr.Code())
			} else {
				require.NoError(t, err)
				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}
		})
	}
}
