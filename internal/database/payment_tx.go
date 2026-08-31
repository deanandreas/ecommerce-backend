package database

import (
	"context"
	"encoding/json"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type orderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

// ConfirmPaymentTx marks a payment as paid and its order as confirmed.
// It only transitions from a pending state so a payment can never be
// confirmed twice or after being cancelled.
func (p *pSQL) ConfirmPaymentTx(ctx context.Context, paymentID, orderID, userID string) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	if err := q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID: paymentID, Status: "paid",
	}); err != nil {
		return err
	}

	if err := q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status: "confirmed", ID: orderID, UserID: userID,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CancelPaymentTx marks a pending payment as cancelled, cancels its order,
// and returns the reserved stock for every item in the order.
func (p *pSQL) CancelPaymentTx(ctx context.Context, paymentID, orderID, userID string) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	order, err := q.GetUnpayedOrder(ctx, userID)
	if err != nil {
		return err
	}

	items, err := parseOrderItems(order.OrderItems)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := q.UpdateProductStock(ctx, db.UpdateProductStockParams{
			Stock: item.Quantity, ID: item.ProductID,
		}); err != nil {
			return err
		}
	}

	if err := q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID: paymentID, Status: "cancelled",
	}); err != nil {
		return err
	}

	if err := q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status: "cancelled", ID: orderID, UserID: userID,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ExpirePendingPaymentsTx cancels every payment still pending after the cutoff
// and returns the reserved stock. It is meant to be called by a background
// job, never by a client.
func (p *pSQL) ExpirePendingPaymentsTx(ctx context.Context, cutoff pgtype.Timestamp) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	payments, err := q.GetExpiredPendingPayments(ctx, cutoff)
	if err != nil {
		return err
	}

	for _, payment := range payments {
		order, err := q.GetUnpayedOrder(ctx, payment.UserID)
		if err != nil {
			continue
		}

		items, err := parseOrderItems(order.OrderItems)
		if err != nil {
			continue
		}

		for _, item := range items {
			if err := q.UpdateProductStock(ctx, db.UpdateProductStockParams{
				Stock: item.Quantity, ID: item.ProductID,
			}); err != nil {
				return err
			}
		}

		if err := q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
			ID: payment.ID, Status: "cancelled",
		}); err != nil {
			return err
		}

		if err := q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
			Status: "cancelled", ID: payment.OrderID, UserID: payment.UserID,
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func parseOrderItems(data interface{}) ([]orderItem, error) {
	var items []orderItem
	rowByte, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rowByte, &items); err != nil {
		return nil, err
	}
	return items, nil
}
