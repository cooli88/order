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

func TestListOrdersHandler(t *testing.T) {
	tests := []struct {
		name         string
		request      *orderv1.ListOrdersRequest
		setupMock    func() *store.MockOrderStore
		wantErr      bool
		wantCode     connect.Code
		validateResp func(t *testing.T, resp *connect.Response[orderv1.ListOrdersResponse])
	}{
		{
			name:    "empty_list",
			request: &orderv1.ListOrdersRequest{},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					ListFunc: func(ctx context.Context) ([]*entity.Order, error) {
						return []*entity.Order{}, nil
					},
				}
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.ListOrdersResponse]) {
				require.NotNil(t, resp)
				require.Empty(t, resp.Msg.Orders)
			},
		},
		{
			name:    "single_order",
			request: &orderv1.ListOrdersRequest{},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					ListFunc: func(ctx context.Context) ([]*entity.Order, error) {
						return []*entity.Order{
							{
								ID:        "order-1",
								UserID:    "user-1",
								Item:      "Item 1",
								Amount:    10.00,
								Status:    entity.OrderStatusNew,
								CreatedAt: time.Now().UTC(),
							},
						}, nil
					},
				}
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.ListOrdersResponse]) {
				require.NotNil(t, resp)
				require.Len(t, resp.Msg.Orders, 1)
				require.Equal(t, "order-1", resp.Msg.Orders[0].Id)
				require.Equal(t, "user-1", resp.Msg.Orders[0].UserId)
				require.Equal(t, "Item 1", resp.Msg.Orders[0].Item)
				require.Equal(t, 10.00, resp.Msg.Orders[0].Amount)
			},
		},
		{
			name:    "multiple_orders",
			request: &orderv1.ListOrdersRequest{},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					ListFunc: func(ctx context.Context) ([]*entity.Order, error) {
						return []*entity.Order{
							{
								ID:        "order-1",
								UserID:    "user-1",
								Item:      "Item 1",
								Amount:    10.00,
								Status:    entity.OrderStatusNew,
								CreatedAt: time.Now().UTC(),
							},
							{
								ID:        "order-2",
								UserID:    "user-2",
								Item:      "Item 2",
								Amount:    20.00,
								Status:    entity.OrderStatusInProgress,
								CreatedAt: time.Now().UTC(),
							},
							{
								ID:        "order-3",
								UserID:    "user-1",
								Item:      "Item 3",
								Amount:    30.00,
								Status:    entity.OrderStatusFinished,
								CreatedAt: time.Now().UTC(),
							},
						}, nil
					},
				}
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *connect.Response[orderv1.ListOrdersResponse]) {
				require.NotNil(t, resp)
				require.Len(t, resp.Msg.Orders, 3)
				require.Equal(t, "order-1", resp.Msg.Orders[0].Id)
				require.Equal(t, "order-2", resp.Msg.Orders[1].Id)
				require.Equal(t, "order-3", resp.Msg.Orders[2].Id)
			},
		},
		{
			name:    "store_error",
			request: &orderv1.ListOrdersRequest{},
			setupMock: func() *store.MockOrderStore {
				return &store.MockOrderStore{
					ListFunc: func(ctx context.Context) ([]*entity.Order, error) {
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
			handler := newListOrdersHandler(mockStore)

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
