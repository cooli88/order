package isolation

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/stretchr/testify/suite"
)

type ListOrdersSuite struct {
	Suite
}

func TestListOrdersSuite(t *testing.T) {
	suite.Run(t, new(ListOrdersSuite))
}

func (s *ListOrdersSuite) TestListOrders_MultipleOrders() {
	s.WithAllure("ListOrders_MultipleOrders", "Verify listing multiple orders returns all created orders")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Create multiple orders
	order1 := s.CreateOrder(ctx, userID, "Product A", 10.00)
	order2 := s.CreateOrder(ctx, userID, "Product B", 20.00)
	order3 := s.CreateOrder(ctx, userID, "Product C", 30.00)

	// List all orders
	resp, err := s.orderClient.ListOrders(ctx, connect.NewRequest(&orderv1.ListOrdersRequest{}))

	s.Require().NoError(err)
	s.Require().NotNil(resp.Msg.Orders)

	// Verify all created orders are in the list
	orderIDs := make(map[string]bool)
	for _, o := range resp.Msg.Orders {
		orderIDs[o.Id] = true
	}

	s.Require().True(orderIDs[order1.Id], "Order 1 should be in the list")
	s.Require().True(orderIDs[order2.Id], "Order 2 should be in the list")
	s.Require().True(orderIDs[order3.Id], "Order 3 should be in the list")
}

func (s *ListOrdersSuite) TestListOrders_ContainsCreatedOrder() {
	s.WithAllure("ListOrders_ContainsCreatedOrder", "Verify a newly created order appears in the list")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Create a unique order
	uniqueItem := "Unique-" + s.GenerateUserID()
	createdOrder := s.CreateOrder(ctx, userID, uniqueItem, 99.99)

	// List orders
	resp, err := s.orderClient.ListOrders(ctx, connect.NewRequest(&orderv1.ListOrdersRequest{}))

	s.Require().NoError(err)

	// Find the created order in the list
	var foundOrder *orderv1.Order
	for _, o := range resp.Msg.Orders {
		if o.Id == createdOrder.Id {
			foundOrder = o
			break
		}
	}

	s.Require().NotNil(foundOrder, "Created order should be found in list")
	s.Require().Equal(userID, foundOrder.UserId)
	s.Require().Equal(uniqueItem, foundOrder.Item)
	s.Require().Equal(99.99, foundOrder.Amount)
}
