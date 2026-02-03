package orders

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/entity"
	"github.com/demo/order/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderHandler(t *testing.T) {
	// Define testData struct locally - stores all test dependencies and results
	type testData struct {
		ctx       context.Context
		t         *testing.T
		handler   *createOrderHandler
		mockStore *store.MockOrderStore
		request   *connect.Request[orderv1.CreateOrderRequest]
		response  *connect.Response[orderv1.CreateOrderResponse]
		err       error
		// Track Create calls for verification
		createCalls []*entity.Order
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
		td := &testData{
			ctx:         context.Background(),
			t:           t,
			createCalls: make([]*entity.Order, 0),
		}

		td.mockStore = &store.MockOrderStore{}

		// Setup default mock behavior - successful create that captures the order
		td.mockStore.CreateFunc = func(_ context.Context, order *entity.Order) error {
			td.createCalls = append(td.createCalls, order)
			return nil
		}

		td.handler = newCreateOrderHandler(td.mockStore)

		return td
	}

	testCases := []testCase{
		// Success scenario
		{
			name: "Should create order successfully with valid input",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "user-123",
					Item:   "Test Item",
					Amount: 100.50,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				require.NotNil(td.t, td.response.Msg.Order)

				// Verify response contains correct data
				assert.NotEmpty(td.t, td.response.Msg.Order.Id, "Order ID should be generated")
				assert.Equal(td.t, "user-123", td.response.Msg.Order.UserId)
				assert.Equal(td.t, "Test Item", td.response.Msg.Order.Item)
				assert.Equal(td.t, 100.50, td.response.Msg.Order.Amount)
				assert.Equal(td.t, string(entity.OrderStatusNew), td.response.Msg.Order.Status)
				assert.NotEmpty(td.t, td.response.Msg.Order.CreatedAt)

				// Verify store was called once
				assert.Len(td.t, td.createCalls, 1, "Store.Create should be called once")

				// Verify order passed to store
				savedOrder := td.createCalls[0]
				assert.NotEmpty(td.t, savedOrder.ID)
				assert.Equal(td.t, "user-123", savedOrder.UserID)
				assert.Equal(td.t, "Test Item", savedOrder.Item)
				assert.Equal(td.t, 100.50, savedOrder.Amount)
				assert.Equal(td.t, entity.OrderStatusNew, savedOrder.Status)
			},
		},

		// Validation error - empty user_id
		{
			name: "Should return InvalidArgument when user_id is empty",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "",
					Item:   "Test Item",
					Amount: 100.50,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInvalidArgument, connect.CodeOf(td.err))
				assert.Len(td.t, td.createCalls, 0, "Store.Create should not be called on validation error")
			},
		},

		// Validation error - empty item
		{
			name: "Should return InvalidArgument when item is empty",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "user-123",
					Item:   "",
					Amount: 100.50,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInvalidArgument, connect.CodeOf(td.err))
				assert.Len(td.t, td.createCalls, 0, "Store.Create should not be called on validation error")
			},
		},

		// Validation error - zero amount
		{
			name: "Should return InvalidArgument when amount is zero",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "user-123",
					Item:   "Test Item",
					Amount: 0,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInvalidArgument, connect.CodeOf(td.err))
				assert.Len(td.t, td.createCalls, 0, "Store.Create should not be called on validation error")
			},
		},

		// Validation error - negative amount
		{
			name: "Should return InvalidArgument when amount is negative",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "user-123",
					Item:   "Test Item",
					Amount: -10.00,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInvalidArgument, connect.CodeOf(td.err))
				assert.Len(td.t, td.createCalls, 0, "Store.Create should not be called on validation error")
			},
		},

		// Store error
		{
			name: "Should return Internal error when store.Create fails",
			given: func(td *testData) {
				td.mockStore.CreateFunc = func(_ context.Context, order *entity.Order) error {
					td.createCalls = append(td.createCalls, order)
					return errors.New("database connection failed")
				}
				td.request = connect.NewRequest(&orderv1.CreateOrderRequest{
					UserId: "user-123",
					Item:   "Test Item",
					Amount: 100.50,
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Nil(td.t, td.response)
				assert.Equal(td.t, connect.CodeInternal, connect.CodeOf(td.err))
				assert.Len(td.t, td.createCalls, 1, "Store.Create should be called once before failing")
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
