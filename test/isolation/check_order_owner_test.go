package isolation

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CheckOrderOwnerSuite struct {
	Suite
}

func TestCheckOrderOwnerSuite(t *testing.T) {
	suite.Run(t, new(CheckOrderOwnerSuite))
}

func (s *CheckOrderOwnerSuite) TestCheckOrderOwner_Success() {
	s.WithAllure("CheckOrderOwner_Success", "Verify owner check succeeds for correct user")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Create order for this user
	order := s.CreateOrder(ctx, userID, "Test Item", 50.00)

	// Check owner with correct user
	resp, err := s.orderClient.CheckOrderOwner(ctx, connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
		OrderId: order.Id,
		UserId:  userID,
	}))

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Msg)
}

func (s *CheckOrderOwnerSuite) TestCheckOrderOwner_PermissionDenied() {
	s.WithAllure("CheckOrderOwner_PermissionDenied", "Verify permission denied for wrong user")

	ctx := context.Background()
	ownerUserID := s.GenerateUserID()
	otherUserID := s.GenerateUserID()

	// Create order for owner
	order := s.CreateOrder(ctx, ownerUserID, "Test Item", 50.00)

	// Check owner with different user
	_, err := s.orderClient.CheckOrderOwner(ctx, connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
		OrderId: order.Id,
		UserId:  otherUserID,
	}))

	s.Require().Error(err)
	var connectErr *connect.Error
	s.Require().ErrorAs(err, &connectErr)
	s.Require().Equal(connect.CodePermissionDenied, connectErr.Code())
}

func (s *CheckOrderOwnerSuite) TestCheckOrderOwner_OrderNotFound() {
	s.WithAllure("CheckOrderOwner_OrderNotFound", "Verify NotFound error for non-existent order")

	ctx := context.Background()
	userID := s.GenerateUserID()

	// Check owner for non-existent order
	nonExistentID := uuid.New().String()
	_, err := s.orderClient.CheckOrderOwner(ctx, connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
		OrderId: nonExistentID,
		UserId:  userID,
	}))

	s.Require().Error(err)
	var connectErr *connect.Error
	s.Require().ErrorAs(err, &connectErr)
	s.Require().Equal(connect.CodeNotFound, connectErr.Code())
}

func (s *CheckOrderOwnerSuite) TestCheckOrderOwner_ValidationErrors() {
	s.WithAllure("CheckOrderOwner_ValidationErrors", "Verify validation errors for empty fields")

	ctx := context.Background()

	testCases := []struct {
		name    string
		orderID string
		userID  string
		wantErr connect.Code
	}{
		{
			name:    "empty_order_id",
			orderID: "",
			userID:  s.GenerateUserID(),
			wantErr: connect.CodeInvalidArgument,
		},
		{
			name:    "empty_user_id",
			orderID: uuid.New().String(),
			userID:  "",
			wantErr: connect.CodeInvalidArgument,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.orderClient.CheckOrderOwner(ctx, connect.NewRequest(&orderv1.CheckOrderOwnerRequest{
				OrderId: tc.orderID,
				UserId:  tc.userID,
			}))

			s.Require().Error(err, "Expected error for %s", tc.name)
			var connectErr *connect.Error
			s.Require().ErrorAs(err, &connectErr)
			s.Require().Equal(tc.wantErr, connectErr.Code())
		})
	}
}
