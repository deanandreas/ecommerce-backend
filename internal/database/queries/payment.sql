-- name: InsertPayment :one
INSERT INTO "payments" ("user_id", "transaction_id", "status", "order_id", "payment_gateway", "amount_in_cent", "currency", "payment_method")
    VALUES ($1, $2, 'pending', $3, $4, $5, $6, $7)
RETURNING
    "id", "user_id", "transaction_id", "status", "order_id", "payment_gateway", "amount_in_cent", "currency", "payment_method", "created_at";

-- name: GetPaymentByOrder :one
SELECT
    *
FROM
    "payments"
WHERE
    "order_id" = $1
    AND "user_id" = $2;

-- name: GetPaymentByTransaction :one
SELECT
    *
FROM
    "payments"
WHERE
    "transaction_id" = $1;

-- name: GetPaymentByID :one
SELECT
    *
FROM
    "payments"
WHERE
    "id" = $1;

-- name: UpdatePaymentStatus :execrows
UPDATE "payments"
SET
    "status" = $1
WHERE
    "id" = $2
    AND "status" = 'pending';

-- name: GetExpiredPendingPayments :many
SELECT
    *
FROM
    "payments"
WHERE
    "status" = 'pending'
    AND "created_at" < $1;

-- name: UpdateOrderStatus :execrows
UPDATE "orders"
SET
    "status" = $1
WHERE
    "id" = $2
    AND "user_id" = $3;
