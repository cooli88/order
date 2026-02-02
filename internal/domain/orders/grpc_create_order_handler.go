package orders

import (
	"context"
	"time"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/entity"
	"github.com/demo/order/internal/store"
	"github.com/google/uuid"
)

type createOrderHandler struct {
	store store.OrderStore
}

func newCreateOrderHandler(store store.OrderStore) *createOrderHandler {
	return &createOrderHandler{store: store}
}

func (h *createOrderHandler) Handle(
	ctx context.Context,
	req *connect.Request[orderv1.CreateOrderRequest],
) (*connect.Response[orderv1.CreateOrderResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}

	order := &entity.Order{
		ID:        uuid.New().String(),
		UserID:    req.Msg.UserId,
		Item:      req.Msg.Item,
		Amount:    req.Msg.Amount,
		Status:    entity.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.store.Create(ctx, order); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&orderv1.CreateOrderResponse{
		Order: entityToProto(order),
	}), nil
}

func (h *createOrderHandler) validate(req *orderv1.CreateOrderRequest) error {
	if req.UserId == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.Item == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.Amount <= 0 {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	return nil
}
