package entity

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusInProgress OrderStatus = "IN_PROGRESS"
	OrderStatusFinished   OrderStatus = "FINISHED"
)

type Order struct {
	ID        string      `db:"id"`
	UserID    string      `db:"user_id"`
	Item      string      `db:"item"`
	Amount    float64     `db:"amount"`
	Status    OrderStatus `db:"status"`
	CreatedAt time.Time   `db:"created_at"`
}
