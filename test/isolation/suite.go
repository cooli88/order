package isolation

import (
	"context"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/contracts/gen/go/order/v1/orderv1connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const (
	defaultBaseURL = "http://localhost:8081"
)

// Suite is the base test suite for isolation tests.
// It provides Connect RPC client for BLACK-BOX testing (no direct DB access).
type Suite struct {
	suite.Suite

	// Connect RPC client for Order Service
	orderClient orderv1connect.OrderServiceClient

	// Base URL for the service
	baseURL string
}

// SetupSuite initializes the test suite with RPC client.
func (s *Suite) SetupSuite() {
	s.baseURL = os.Getenv("ORDER_SERVICE_URL")
	if s.baseURL == "" {
		s.baseURL = defaultBaseURL
	}

	// Create HTTP client for Connect RPC
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Initialize Order Service client
	s.orderClient = orderv1connect.NewOrderServiceClient(httpClient, s.baseURL)
}

// WithAllure logs test metadata for Allure reporting.
func (s *Suite) WithAllure(name, description string) {
	s.T().Logf("Allure Test: %s", name)
	s.T().Logf("Description: %s", description)
}

// CreateOrder is a helper to create an order via RPC and return the response.
func (s *Suite) CreateOrder(ctx context.Context, userID, item string, amount float64) *orderv1.Order {
	resp, err := s.orderClient.CreateOrder(ctx, connect.NewRequest(&orderv1.CreateOrderRequest{
		UserId: userID,
		Item:   item,
		Amount: amount,
	}))
	s.Require().NoError(err, "Failed to create order")
	s.Require().NotNil(resp.Msg.Order, "Order should not be nil")
	return resp.Msg.Order
}

// GenerateUserID creates a unique user ID for test isolation.
func (s *Suite) GenerateUserID() string {
	return uuid.New().String()
}
