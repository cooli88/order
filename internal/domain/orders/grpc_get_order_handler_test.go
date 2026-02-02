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

func TestGetOrderHandler(t *testing.T) {
	tests := []struct {
		name         string
		request      *orderv1.GetOrderRequest
		setupMock    func() *store.MockOrderStore
		wantErr      bool
		wantCode     connect.Code
		validateResp func(t *testing.T, resp *connect.Response[orderv1.GetOrderResponse])
	}{
		{
			name: "success",
			request: &orderv1.GetOrderRequest{
				Id: "order-123",
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
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.GetOrderResponse]) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg.Order)
				require.Equal(t, "order-123", resp.Msg.Order.Id)
				require.Equal(t, "user-456", resp.Msg.Order.UserId)
				require.Equal(t, "Test Item", resp.Msg.Order.Item)
				require.Equal(t, 49.99, resp.Msg.Order.Amount)
			},
		},
		{
			name: "empty_id",
			request: &orderv1.GetOrderRequest{
				Id: "",
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "not_found",
			request: &orderv1.GetOrderRequest{
				Id: "non-existent-order",
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
			name: "store_error",
			request: &orderv1.GetOrderRequest{
				Id: "order-123",
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
			handler := newGetOrderHandler(mockStore)

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
