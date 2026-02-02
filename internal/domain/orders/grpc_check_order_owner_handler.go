package orders

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/store"
)

type checkOrderOwnerHandler struct {
	store store.OrderStore
}

func newCheckOrderOwnerHandler(store store.OrderStore) *checkOrderOwnerHandler {
	return &checkOrderOwnerHandler{store: store}
}

func (h *checkOrderOwnerHandler) Handle(
	ctx context.Context,
	req *connect.Request[orderv1.CheckOrderOwnerRequest],
) (*connect.Response[orderv1.CheckOrderOwnerResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}

	order, err := h.store.Get(ctx, req.Msg.OrderId)
	if err != nil {
		if errors.Is(err, store.ErrOrderNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if order.UserID != req.Msg.UserId {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("order does not belong to user"))
	}

	return connect.NewResponse(&orderv1.CheckOrderOwnerResponse{}), nil
}

func (h *checkOrderOwnerHandler) validate(req *orderv1.CheckOrderOwnerRequest) error {
	if req.OrderId == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.UserId == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	return nil
}
