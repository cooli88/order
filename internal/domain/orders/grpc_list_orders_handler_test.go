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

func TestListOrdersHandler(t *testing.T) {
	// Define testData struct locally - Order Service specific
	type testData struct {
		ctx       context.Context
		t         *testing.T
		handler   *listOrdersHandler
		mockStore *store.MockOrderStore
		request   *connect.Request[orderv1.ListOrdersRequest]
		response  *connect.Response[orderv1.ListOrdersResponse]
		err       error

		// Helper fields for test setup
		listCalled bool
	}

	// Define testCase struct locally - GWT pattern is MANDATORY
	type testCase struct {
		name  string
		given func(*testData)
		when  func(*testData)
		then  func(*testData)
	}

	// Setup function creates isolated test data for each test case
	setupTestData := func(t *testing.T) *testData {
		mockStore := &store.MockOrderStore{}

		// Setup default mock behavior (empty list)
		mockStore.ListFunc = func(ctx context.Context) ([]*entity.Order, error) {
			return []*entity.Order{}, nil
		}

		handler := newListOrdersHandler(mockStore)

		return &testData{
			ctx:       context.Background(),
			t:         t,
			handler:   handler,
			mockStore: mockStore,
			request:   connect.NewRequest(&orderv1.ListOrdersRequest{}),
		}
	}

	testCases := []testCase{
		// Success scenario: empty list
		{
			name: "Should return empty list when no orders exist",
			given: func(td *testData) {
				td.mockStore.ListFunc = func(ctx context.Context) ([]*entity.Order, error) {
					td.listCalled = true
					return []*entity.Order{}, nil
				}
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				assert.Empty(td.t, td.response.Msg.Orders)
				assert.True(td.t, td.listCalled, "Store.List should be called")
			},
		},

		// Success scenario: single order
		{
			name: "Should return single order when one order exists",
			given: func(td *testData) {
				createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				td.mockStore.ListFunc = func(ctx context.Context) ([]*entity.Order, error) {
					td.listCalled = true
					return []*entity.Order{
						{
							ID:        "order-001",
							UserID:    "user-123",
							Item:      "Widget",
							Amount:    99.99,
							Status:    entity.OrderStatusNew,
							CreatedAt: createdAt,
						},
					}, nil
				}
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				require.Len(td.t, td.response.Msg.Orders, 1)
				assert.True(td.t, td.listCalled, "Store.List should be called")

				order := td.response.Msg.Orders[0]
				assert.Equal(td.t, "order-001", order.Id)
				assert.Equal(td.t, "user-123", order.UserId)
				assert.Equal(td.t, "Widget", order.Item)
				assert.Equal(td.t, 99.99, order.Amount)
				assert.Equal(td.t, "NEW", order.Status)
				assert.Equal(td.t, "2024-01-15T10:30:00Z", order.CreatedAt)
			},
		},

		// Success scenario: multiple orders
		{
			name: "Should return multiple orders when several orders exist",
			given: func(td *testData) {
				createdAt1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				createdAt2 := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)
				createdAt3 := time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC)
				td.mockStore.ListFunc = func(ctx context.Context) ([]*entity.Order, error) {
					td.listCalled = true
					return []*entity.Order{
						{
							ID:        "order-003",
							UserID:    "user-789",
							Item:      "Gadget Pro",
							Amount:    299.99,
							Status:    entity.OrderStatusFinished,
							CreatedAt: createdAt3,
						},
						{
							ID:        "order-002",
							UserID:    "user-456",
							Item:      "Super Gadget",
							Amount:    199.99,
							Status:    entity.OrderStatusInProgress,
							CreatedAt: createdAt2,
						},
						{
							ID:        "order-001",
							UserID:    "user-123",
							Item:      "Widget",
							Amount:    99.99,
							Status:    entity.OrderStatusNew,
							CreatedAt: createdAt1,
						},
					}, nil
				}
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				require.Len(td.t, td.response.Msg.Orders, 3)
				assert.True(td.t, td.listCalled, "Store.List should be called")

				// Verify first order (most recent)
				order1 := td.response.Msg.Orders[0]
				assert.Equal(td.t, "order-003", order1.Id)
				assert.Equal(td.t, "user-789", order1.UserId)
				assert.Equal(td.t, "Gadget Pro", order1.Item)
				assert.Equal(td.t, 299.99, order1.Amount)
				assert.Equal(td.t, "FINISHED", order1.Status)

				// Verify second order
				order2 := td.response.Msg.Orders[1]
				assert.Equal(td.t, "order-002", order2.Id)
				assert.Equal(td.t, "user-456", order2.UserId)
				assert.Equal(td.t, "Super Gadget", order2.Item)
				assert.Equal(td.t, 199.99, order2.Amount)
				assert.Equal(td.t, "IN_PROGRESS", order2.Status)

				// Verify third order (oldest)
				order3 := td.response.Msg.Orders[2]
				assert.Equal(td.t, "order-001", order3.Id)
				assert.Equal(td.t, "user-123", order3.UserId)
				assert.Equal(td.t, "Widget", order3.Item)
				assert.Equal(td.t, 99.99, order3.Amount)
				assert.Equal(td.t, "NEW", order3.Status)
			},
		},

		// Error scenario: store returns error
		{
			name: "Should return Internal error when store fails",
			given: func(td *testData) {
				td.mockStore.ListFunc = func(ctx context.Context) ([]*entity.Order, error) {
					td.listCalled = true
					return nil, errors.New("database connection lost")
				}
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInternal, connect.CodeOf(td.err))
				assert.Contains(td.t, td.err.Error(), "database connection lost")
				assert.True(td.t, td.listCalled, "Store.List should be called")
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
