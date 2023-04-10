package queue

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStatusTask struct {
	OrderID  string
	StatusID int
}

type OrderStatusNotifier interface {
	Acquire(ctx context.Context) (*OrderStatusTask, error)
	UpdateStatusAndRelease(ctx context.Context, orderID string, newStatusID int) error
	Add(ctx context.Context, orderID string) error
	Delete(ctx context.Context, orderID string) error
}

type OrderStatusQueue struct {
	conn *pgxpool.Pool
}

func NewOrderStatusQueue(conn *pgxpool.Pool) *OrderStatusQueue {
	return &OrderStatusQueue{conn: conn}
}

func (q *OrderStatusQueue) Acquire(ctx context.Context) (*OrderStatusTask, error) {
	var orderID string
	var statusID int
	err := q.conn.QueryRow(ctx, aquireQuery).Scan(&orderID, &statusID)
	if err != nil {
		return nil, err
	}
	return &OrderStatusTask{OrderID: orderID, StatusID: statusID}, nil
}

func (q *OrderStatusQueue) UpdateStatusAndRelease(ctx context.Context, orderID string, newStatusID int) error {
	_, err := q.conn.Exec(ctx, updateAndReleaseQuery, newStatusID, orderID)
	return err
}

func (q *OrderStatusQueue) Add(ctx context.Context, orderID string) error {
	_, err := q.conn.Exec(ctx, addQuery, orderID)
	return err
}

func (q *OrderStatusQueue) Delete(ctx context.Context, orderID string) error {
	_, err := q.conn.Exec(ctx, deleteQuery, orderID)
	return err
}
