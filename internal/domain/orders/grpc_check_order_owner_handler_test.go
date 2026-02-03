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

func TestCheckOrderOwnerHandler(t *testing.T) {
	// testData contains all data needed for a single test case
	type testData struct {
		ctx       context.Context
		t         *testing.T
		handler   *checkOrderOwnerHandler
		mockStore *store.MockOrderStore
		request   *connect.Request[orderv1.CheckOrderOwnerRequest]
		response  *connect.Response[orderv1.CheckOrderOwnerResponse]
		err       error
	}

	// testCase defines the GWT structure for each test
	type testCase struct {
		name  string
		given func(*testData)
		when  func(*testData)
		then  func(*testData)
	}

	// setupTestData creates isolated test data for each test case
	setupTestData := func(t *testing.T) *testData {
		mockStore := &store.MockOrderStore{}

		// Default mock behavior: return a valid order
		mockStore.GetFunc = func(_ context.Context, id string) (*entity.Order, error) {
			return &entity.Order{
				ID:        id,
				UserID:    "user-123",
				Item:      "Test Item",
				Amount:    100.00,
				Status:    entity.OrderStatusNew,
				CreatedAt: time.Now(),
			}, nil
		}

		handler := newCheckOrderOwnerHandler(mockStore)

		return &testData{
			ctx:       context.Background(),
			t:         t,
			handler:   handler,
			mockStore: mockStore,
		}
	}

	testCases := []testCase{
		// Success scenario: owner matches
		{
			name: "Should return success when user is the owner of the order",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(_ context.Context, id string) (*entity.Order, error) {
					return &entity.Order{
						ID:        id,
						UserID:    "user-123",
						Item:      "Test Item",
						Amount:    100.00,
						Status:    entity.OrderStatusNew,
						CreatedAt: time.Now(),
					}, nil
				}
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "order-456",
					UserId:  "user-123",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.NoError(td.t, td.err)
				require.NotNil(td.t, td.response)
				// Verify empty response (success means user is the owner)
				assert.Equal(td.t, &orderv1.CheckOrderOwnerResponse{}, td.response.Msg)
			},
		},

		// Validation error: empty order_id
		{
			name: "Should return InvalidArgument when order_id is empty",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "",
					UserId:  "user-123",
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

		// Validation error: empty user_id
		{
			name: "Should return InvalidArgument when user_id is empty",
			given: func(td *testData) {
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "order-456",
					UserId:  "",
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

		// NotFound: store returns ErrOrderNotFound
		{
			name: "Should return NotFound when order does not exist",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(_ context.Context, _ string) (*entity.Order, error) {
					return nil, store.ErrOrderNotFound
				}
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "non-existent-order",
					UserId:  "user-123",
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

		// PermissionDenied: user_id doesn't match order's user_id
		{
			name: "Should return PermissionDenied when user is not the owner",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(_ context.Context, id string) (*entity.Order, error) {
					return &entity.Order{
						ID:        id,
						UserID:    "owner-user-789",
						Item:      "Test Item",
						Amount:    100.00,
						Status:    entity.OrderStatusNew,
						CreatedAt: time.Now(),
					}, nil
				}
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "order-456",
					UserId:  "different-user-123",
				})
			},
			when: func(td *testData) {
				td.response, td.err = td.handler.Handle(td.ctx, td.request)
			},
			then: func(td *testData) {
				require.Error(td.t, td.err)
				assert.Equal(td.t, connect.CodePermissionDenied, connect.CodeOf(td.err))
				assert.Contains(td.t, td.err.Error(), "order does not belong to user")
				assert.Nil(td.t, td.response)
			},
		},

		// Internal: store.Get() returns other error
		{
			name: "Should return Internal error when store returns unexpected error",
			given: func(td *testData) {
				td.mockStore.GetFunc = func(_ context.Context, _ string) (*entity.Order, error) {
					return nil, errors.New("database connection failed")
				}
				td.request = connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
					OrderId: "order-456",
					UserId:  "user-123",
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
