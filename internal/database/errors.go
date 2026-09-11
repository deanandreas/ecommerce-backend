package database

import "errors"

var (
	ErrNotDeleted        = errors.New("not deleted")
	ErrNotUpdated        = errors.New("not updated")
	ErrInvalidQuantity   = errors.New("to many quantity")
	ErrInvalidType       = errors.New("invalid file type")
	ErrNoFile            = errors.New("no file found")
	ErrNoItem            = errors.New("no item found")
	ErrPaymentNotPending = errors.New("payment is not pending")
)
