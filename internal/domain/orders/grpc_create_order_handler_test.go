package orders

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/entity"
	"github.com/demo/order/internal/store"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderHandler(t *testing.T) {
	tests := []struct {
		name         string
		request      *orderv1.CreateOrderRequest
		setupMock    func() *store.MockOrderStore
		wantErr      bool
		wantCode     connect.Code
		validateResp func(t *testing.T, resp *connect.Response[orderv1.CreateOrderResponse])
	}{
		{
			name: "success",
			request: &orderv1.CreateOrderRequest{
				UserId: "user-123",
				Item:   "Test Item",
				Amount: 99.99,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					CreateFunc: func(ctx context.Context, order *entity.Order) error {
						return nil
					},
				}
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.CreateOrderResponse]) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg.Order)
				require.NotEmpty(t, resp.Msg.Order.Id)
				require.Equal(t, "user-123", resp.Msg.Order.UserId)
				require.Equal(t, "Test Item", resp.Msg.Order.Item)
				require.Equal(t, 99.99, resp.Msg.Order.Amount)
			},
		},
		{
			name: "empty_user_id",
			request: &orderv1.CreateOrderRequest{
				UserId: "",
				Item:   "Test Item",
				Amount: 99.99,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "empty_item",
			request: &orderv1.CreateOrderRequest{
				UserId: "user-123",
				Item:   "",
				Amount: 99.99,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "zero_amount",
			request: &orderv1.CreateOrderRequest{
				UserId: "user-123",
				Item:   "Test Item",
				Amount: 0,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "negative_amount",
			request: &orderv1.CreateOrderRequest{
				UserId: "user-123",
				Item:   "Test Item",
				Amount: -10.00,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{}
			},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "store_error",
			request: &orderv1.CreateOrderRequest{
				UserId: "user-123",
				Item:   "Test Item",
				Amount: 99.99,
			},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					CreateFunc: func(ctx context.Context, order *entity.Order) error {
						return errors.New("database connection failed")
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
			handler := newCreateOrderHandler(mockStore)

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
