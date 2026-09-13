package database

import (
	"context"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

func (p *pSQL) InsertUserTx(ctx context.Context, arg UserData) (*db.GetUserByIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	userID, err := q.InsertUser(ctx, arg.InsertUserParams)
	if err != nil {
		return nil, err
	}

	for _, addr := range arg.Address {
		addr.UserID = userID
		if err := q.InsertAddress(ctx, addr); err != nil {
			return nil, err
		}
	}

	user, err := q.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *pSQL) InsertCartTx(ctx context.Context, arg CreateCart) (*db.CartItem, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	cartID, err := q.InsertCart(ctx, arg.UserID)
	if err != nil {
		return nil, err
	}
	arg.CartID = cartID

	itemID, err := q.InsertCartItem(ctx, arg.InsertCartItemParams)
	if err != nil {
		return nil, err
	}

	cartItem, err := q.GetCartItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	quantity, err := q.GetProductStock(ctx, cartItem.ProductID)
	if err != nil {
		return nil, err
	}

	if quantity < cartItem.Quantity {
		return nil, ErrInvalidQuantity
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &cartItem, nil
}

func (p *pSQL) UpdateCartItemTx(ctx context.Context, arg db.UpdateCartItemQuantityParams) (*db.CartItem, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	rows, err := q.UpdateCartItemQuantity(ctx, arg)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrNotUpdated
	}

	cartItem, err := q.GetCartItemByID(ctx, arg.ID)
	if err != nil {
		return nil, err
	}

	quantity, err := q.GetProductStock(ctx, cartItem.ProductID)
	if err != nil {
		return nil, err
	}

	if quantity < cartItem.Quantity {
		return nil, ErrInvalidQuantity
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &cartItem, nil
}

func (p *pSQL) UpdateDefaultAddressTx(ctx context.Context, arg db.UpdateDefaultAddressParams) (*db.GetUserByIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	err = q.DesableDefaultAddres(ctx, arg.UserID)
	if err != nil {
		return nil, err
	}

	row, err := q.UpdateDefaultAddress(ctx, arg)
	if err != nil {
		return nil, err
	} else if row == 0 {
		return nil, ErrNotUpdated
	}

	user, err := q.GetUserByID(ctx, arg.UserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}
