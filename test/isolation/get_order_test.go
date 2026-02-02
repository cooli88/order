package isolation

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type GetOrderSuite struct {
	Suite
}

func TestGetOrderSuite(t *testing.T) {
	suite.Run(t, new(GetOrderSuite))
}

func (s *GetOrderSuite) TestGetOrder_AfterCreate() {
	s.WithAllure("GetOrder_AfterCreate", "Verify order can be retrieved after creation")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Create order first
	createdOrder := s.CreateOrder(ctx, userID, "Laptop", 1299.99)

	// Get order by ID
	resp, err := s.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{
		Id: createdOrder.Id,
	}))

	s.Require().NoError(err)
	s.Require().NotNil(resp.Msg.Order)

	// Verify all fields match
	order := resp.Msg.Order
	s.Require().Equal(createdOrder.Id, order.Id)
	s.Require().Equal(userID, order.UserId)
	s.Require().Equal("Laptop", order.Item)
	s.Require().Equal(1299.99, order.Amount)
}

func (s *GetOrderSuite) TestGetOrder_NotFound() {
	s.WithAllure("GetOrder_NotFound", "Verify NotFound error for non-existent order")

	ctx := context.Background()

	// Try to get a non-existent order with random UUID
	nonExistentID := uuid.New().String()

	_, err := s.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{
		Id: nonExistentID,
	}))

	s.Require().Error(err)
	var connectErr *connect.Error
	s.Require().ErrorAs(err, &connectErr)
	s.Require().Equal(connect.CodeNotFound, connectErr.Code())
}

func (s *GetOrderSuite) TestGetOrder_EmptyID() {
	s.WithAllure("GetOrder_EmptyID", "Verify InvalidArgument error for empty order ID")

	ctx := context.Background()

	_, err := s.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{
		Id: "",
	}))

	s.Require().Error(err)
	var connectErr *connect.Error
	s.Require().ErrorAs(err, &connectErr)
	s.Require().Equal(connect.CodeInvalidArgument, connectErr.Code())
}
