-- name: InsertCart :one
INSERT INTO "carts"("user_id") VALUES ($1)
ON CONFLICT ("user_id") DO UPDATE SET "user_id" = EXCLUDED."user_id"
RETURNING "id";

-- name: InsertCartItem :one
INSERT INTO "cart_items"("cart_id", "product_id", "quantity")
VALUES ($1, $2, $3) ON CONFLICT ("cart_id", "product_id")
DO UPDATE SET "quantity" = EXCLUDED."quantity" + "cart_items"."quantity"
RETURNING "id";

-- name: GetUserCart :one
SELECT
  c."id",
  COALESCE (
    (
      SELECT JSON_AGG(
        JSON_BUILD_OBJECT(
          'id', i."id",
          'product_id', p."id",
          'title', p."title",
          'price_in_cent', p."price_in_cent",
          'quantity', i."quantity"
        )
      )
      FROM "cart_items" i LEFT JOIN "products" p ON p."id" = i."product_id"
      WHERE i."cart_id" = c."id"
    ),'[]'::JSON
  ) AS "items"
FROM "carts" c WHERE c."user_id" = $1;

-- name: GetCartItemByID :one
SELECT * FROM "cart_items" WHERE "id" = $1;

-- name: UpdateCartItemQuantity :execrows
UPDATE "cart_items" i SET "quantity" = $1, "updated_at" = NOW()
FROM "carts" c WHERE i."id" = $2 AND c."id" = i."cart_id" AND c."user_id" = $3;

-- name: DeleteCart :execrows
DELETE FROM "carts" WHERE "id" = $1 AND "user_id" = $2;

-- name: DeleteCartItem :execrows
DELETE FROM "cart_items" i 
USING "carts" c 
WHERE i."id" = $1 AND c."id" = i."cart_id" AND c."user_id" = $2;

-- name: DeleteAllCartItem :execrows
DELETE FROM "cart_items" i USING "carts" c
WHERE i."cart_id" = $1 AND c."id" = i."cart_id" AND c."user_id" = $2;

