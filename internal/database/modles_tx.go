package database

import (
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type ProductData struct {
	db.InsertProductParams
	Category db.InsertCategoryParams `json:"category"`
}

type UpdateProductData struct {
	db.UpdateProductParams
	Category db.InsertCategoryParams `json:"category"`
}

type UserData struct {
	db.InsertUserParams
	Password string                   `json:"password"`
	Address  []db.InsertAddressParams `json:"address"`
}

type CreateCart struct {
	UserID string
	db.InsertCartItemParams
}
