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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrderHandler(t *testing.T) {
	// testData holds all data needed for each test case
	type testData struct {
		ctx       context.Context
		t         *testing.T
		handler   *getOrderHandler
		mockStore *store.MockOrderStore
		request   *connect.Request[orderv1.GetOrderRequest]
		response  *connect.Response[orderv1.GetOrderResponse]
		err       error
	}

	// testCase defines GWT structure for each test scenario
	type testCase struct {
		name  string
		given func(*testData)
		when  func(*testData)
		then  func(*testData)
	}

	// setupTestData creates isolated test data for each test case
	setupTestData := func(t *testing.T) *testData {
		mockStore := &store.MockOrderStore{}

		handler := newGetOrderHandler(mockStore)

		return &testData{
			ctx:       context.Background(),
			t:         t,
			handler:   handler,
			mockStore: mockStore,
		}
	}

	testCases := []testCase{
		{
			name: "Should return order successfully when order exists",
			given: func(td *testData) {
				expectedOrder := &entity.Order{
					ID:        "order-123",
					UserID:    "user-456",
					Item:      "Test Item",
					Amount:    99.99,
					Status:    entity.OrderStatusNew,
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				}
				td.mockStore.GetFunc = func(ctx context.Context, id string) (*entity.Order, error) {
					if id == "order-123" {
						return expectedOrder, nil
					}
					return nil, store.ErrOrderNotFound
				}
				td.request = connect.NewRequest(&orderv1.GetOrderRequest{
					Id: "order-123",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				require.NotNil(td.t, td.response.Msg.Order)

				order := td.response.Msg.Order
				assert.Equal(td.t, "order-123", order.Id)
				assert.Equal(td.t, "user-456", order.UserId)
				assert.Equal(td.t, "Test Item", order.Item)
				assert.Equal(td.t, 99.99, order.Amount)
				assert.Equal(td.t, "NEW", order.Status)
				assert.Equal(td.t, "2024-01-15T10:30:00Z", order.CreatedAt)
			},
		},
		{
			name: "Should return InvalidArgument when id is empty",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.GetOrderRequest{
					Id: "",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Equal(td.t, connect.CodeInvalidArgument, connect.CodeOf(td.err))
				assert.Nil(td.t, td.response)
			},
		},
		{
			name: "Should return NotFound when order does not exist",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(ctx context.Context, id string) (*entity.Order, error) {
					return nil, store.ErrOrderNotFound
				}
				td.request = connect.NewRequest(&orderv1.GetOrderRequest{
					Id: "non-existent-order",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Equal(td.t, connect.CodeNotFound, connect.CodeOf(td.err))
				assert.Nil(td.t, td.response)
			},
		},
		{
			name: "Should return Internal error when store returns unexpected error",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(ctx context.Context, id string) (*entity.Order, error) {
					return nil, errors.New("database connection failed")
				}
				td.request = connect.NewRequest(&orderv1.GetOrderRequest{
					Id: "order-123",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Equal(td.t, connect.CodeInternal, connect.CodeOf(td.err))
				assert.Nil(td.t, td.response)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td := setupTestData(t)
			td.t = t
			tc.given(td)
			tc.when(td)
			tc.then(td)
		})
	}
}
