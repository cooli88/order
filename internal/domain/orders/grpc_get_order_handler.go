package orders

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/store"
)

type getOrderHandler struct {
	store store.OrderStore
}

func newGetOrderHandler(store store.OrderStore) *getOrderHandler {
	return &getOrderHandler{store: store}
}

func (h *getOrderHandler) Handle(
	ctx context.Context,
	req *connect.Request[orderv1.GetOrderRequest],
) (*connect.Response[orderv1.GetOrderResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}

	order, err := h.store.Get(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, store.ErrOrderNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&orderv1.GetOrderResponse{
		Order: entityToProto(order),
	}), nil
}

func (h *getOrderHandler) validate(req *orderv1.GetOrderRequest) error {
	if req.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	return nil
}
