-- name: InsertOrder :one
INSERT INTO "orders" ("user_id", "address_id", "ship_date", "ship_price_in_cent")
    VALUES ($1, $2, $3, $4)
RETURNING
    "id";

-- name: InsertOrderItem :exec
INSERT INTO "order_items" ("order_id", "product_id", "quantity")
    VALUES ($1, $2, $3);

-- name: GetAllUserOrders :many
SELECT
    o."id",
    o."ship_date",
    o."status",
    o."ship_price_in_cent",
    o."created_at",
    COALESCE((
        SELECT
            ROW_TO_JSON(a_sub)
        FROM (
            SELECT
                a."street_line_1", a."street_line_2", a."postal_code", a."state", a."city", a."country"
            FROM "addresses" a
            WHERE
                a."id" = o."address_id") a_sub), '{}'::json) AS "address",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', i."id", 'title', p."title", 'stock', p."stock", 'quantity', i.quantity))
        FROM "order_items" i
    LEFT JOIN "products" p ON p."id" = i."product_id"
    WHERE
        i."order_id" = o."id"), '[]'::json) AS "order_items"
FROM
    "orders" o
WHERE
    o."user_id" = $1;

-- name: GetUserOrder :one
SELECT
    o."id",
    o."ship_date",
    o."status",
    o."ship_price_in_cent",
    o."created_at",
    COALESCE((
        SELECT
            ROW_TO_JSON(a_sub)
        FROM (
            SELECT
                a."street_line_1", a."street_line_2", a."postal_code", a."state", a."city", a."country"
            FROM "addresses" a
            WHERE
                a."id" = o."address_id") a_sub), '{}'::json) AS "address",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', i."id", 'title', p."title", 'stock', p."stock", 'quantity', i.quantity))
        FROM "order_items" i
    LEFT JOIN "products" p ON p."id" = i."product_id"
    WHERE
        i."order_id" = o."id"), '[]'::json) AS "order_items"
FROM
    "orders" o
WHERE
    o."user_id" = $1
    AND o."id" = $2;

-- name: GetUnpayedOrder :one
SELECT
    o."id",
    o."ship_price_in_cent",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', i."id", 'product_id', p."id", 'price_in_cent', p."price_in_cent", 'quantity', i.quantity))
        FROM "order_items" i
    LEFT JOIN "products" p ON p."id" = i."product_id"
    WHERE
        i."order_id" = o."id"), '[]'::json) AS "order_items"
FROM
    "orders" o
WHERE
    o."user_id" = $1
    AND "status" = 'pending';

-- name: DeleteUserOrder :exec
DELETE FROM "orders"
WHERE "id" = $1
    AND "user_id" = $2;

