package isolation

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/stretchr/testify/suite"
)

type CreateOrderSuite struct {
	Suite
}

func TestCreateOrderSuite(t *testing.T) {
	suite.Run(t, new(CreateOrderSuite))
}

func (s *CreateOrderSuite) TestCreateOrder_Success() {
	s.WithAllure("CreateOrder_Success", "Verify successful order creation with valid data")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Create order
	order := s.CreateOrder(ctx, userID, "Test Product", 199.99)

	// Verify response
	s.Require().NotEmpty(order.Id, "Order ID should be generated")
	s.Require().Equal(userID, order.UserId, "User ID should match")
	s.Require().Equal("Test Product", order.Item, "Item should match")
	s.Require().Equal(199.99, order.Amount, "Amount should match")

	// Verify order can be retrieved
	getResp, err := s.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{
		Id: order.Id,
	}))
	s.Require().NoError(err)
	s.Require().Equal(order.Id, getResp.Msg.Order.Id)
	s.Require().Equal(userID, getResp.Msg.Order.UserId)
}

func (s *CreateOrderSuite) TestCreateOrder_ValidationErrors() {
	s.WithAllure("CreateOrder_ValidationErrors", "Verify validation errors for invalid input")

	ctx := context.Background()

	testCases := []struct {
		name    string
		userID  string
		item    string
		amount  float64
		wantErr connect.Code
	}{
		{
			name:    "empty_user_id",
			userID:  "",
			item:    "Test Item",
			amount:  10.00,
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "empty_item",
			userID:  s.GenerateUserID(),
			item:    "",
			amount:  10.00,
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "zero_amount",
			userID:  s.GenerateUserID(),
			item:    "Test Item",
			amount:  0,
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "negative_amount",
			userID:  s.GenerateUserID(),
			item:    "Test Item",
			amount:  -50.00,
			wantErr: connect.CodeInvalidArgument,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.orderClient.CreateOrder(ctx, connect.NewRequest(&orderv1.CreateOrderRequest{
				UserId: tc.userID,
				Item:   tc.item,
				Amount: tc.amount,
			}))

			s.Require().Error(err, "Expected error for %s", tc.name)
			var connectErr *connect.Error
			s.Require().ErrorAs(err, &connectErr)
			s.Require().Equal(tc.wantErr, connectErr.Code(), "Expected %v error code", tc.wantErr)
		})
	}
}
