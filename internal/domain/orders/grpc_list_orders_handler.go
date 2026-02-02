package orders

import (
	"context"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/store"
)

type listOrdersHandler struct {
	store store.OrderStore
}

func newListOrdersHandler(store store.OrderStore) *listOrdersHandler {
	return &listOrdersHandler{store: store}
}

func (h *listOrdersHandler) Handle(
	ctx context.Context,
	req *connect.Request[orderv1.ListOrdersRequest],
) (*connect.Response[orderv1.ListOrdersResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}

	orders, err := h.store.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoOrders := make([]*orderv1.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = entityToProto(o)
	}

	return connect.NewResponse(&orderv1.ListOrdersResponse{
		Orders: protoOrders,
	}), nil
}

func (h *listOrdersHandler) validate(_ *orderv1.ListOrdersRequest) error {
	return nil
}
