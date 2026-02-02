package orders

import (
	"time"

	orderv1 "github.com/demo/contracts/gen/go/order/v1"
	"github.com/demo/order/internal/entity"
)

func entityToProto(e *entity.Order) *orderv1.Order {
	return &orderv1.Order{
		Id:        e.ID,
		UserId:    e.UserID,
		Item:      e.Item,
		Amount:    e.Amount,
		Status:    string(e.Status),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}
