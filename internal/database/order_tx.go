package database

import (
	"context"
	"encoding/json"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

func (p *pSQL) InsertOrderTx(ctx context.Context, arg db.InsertOrderParams) (*db.GetUserOrderRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	cart, err := q.GetUserCart(ctx, arg.UserID)
	if err != nil {
		return nil, err
	}

	var items []struct {
		ProductID   string `json:"product_id"`
		PriceInCent int32  `json:"price_in_cent"`
		Quantity    int32  `json:"quantity"`
	}
	rowByte, err := json.Marshal(cart.Items)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rowByte, &items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, ErrNoItem
	}

	for _, item := range items {
		stock, err := q.GetProductStock(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		if stock < item.Quantity {
			return nil, ErrInvalidQuantity
		}

		if err := q.UpdateProductStock(ctx,
			db.UpdateProductStockParams{
				Stock: stock - item.Quantity, ID: item.ProductID,
			}); err != nil {
			return nil, err
		}
	}

	if arg.AddressID == "" {
		addresses, err := q.GetAddressByUserID(ctx, arg.UserID)
		if err != nil {
			return nil, err
		}

		for _, address := range addresses {
			if address.IsDefault {
				arg.AddressID = address.ID
			}
		}
	}

	orderID, err := q.InsertOrder(ctx, arg)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if err := q.InsertOrderItem(ctx,
			db.InsertOrderItemParams{
				OrderID: orderID, ProductID: item.ProductID, Quantity: item.Quantity,
			}); err != nil {
			return nil, err
		}
	}

	order, err := q.GetUserOrder(ctx, db.GetUserOrderParams{UserID: arg.UserID, ID: orderID})
	if err != nil {
		return nil, err
	}

	row, err := q.DeleteCart(ctx, db.DeleteCartParams{ID: cart.ID, UserID: arg.UserID})
	if err != nil {
		return nil, err
	} else if row == 0 {
		return nil, ErrNotDeleted
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &order, nil
}

func (p *pSQL) DeleteCancelledOrderTx(ctx context.Context, userID string) error {
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

	var items []struct {
		ProductID string `json:"product_id"`
		Quantity  uint32 `json:"quantity"`
	}
	rowByte, err := json.Marshal(order.OrderItems)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rowByte, &items); err != nil {
		return err
	}

	for _, item := range items {
		if err := q.UpdateProductStock(ctx, db.UpdateProductStockParams{Stock: int32(item.Quantity), ID: item.ProductID}); err != nil {
			return err
		}
	}

	err = q.DeleteUserOrder(ctx, db.DeleteUserOrderParams{UserID: userID, ID: order.ID})
	if err != nil {
		return nil
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
