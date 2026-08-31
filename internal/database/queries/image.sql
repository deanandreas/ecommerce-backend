-- name: InsertProductImage :exec 
INSERT INTO "product_images" ("product_id", "image_url", "is_default") 
VALUES ($1, $2, $3);

-- name: GetProductImages :one
SELECT "image_url", "is_default" FROM "product_images" 
WHERE "product_id" = $1 AND "id" = $2;

-- name: UpdateDefaultImage :execrows
UPDATE "product_images" SET "is_default" = $1
WHERE "id" = $2;

-- name: GetDefualtImageID :one
SELECT "id" FROM "product_images" WHERE "product_id" = $1 AND "is_default" = TRUE;

-- name: DeleteProductImage :exec
DELETE FROM "product_images" i
USING "products" p
WHERE i."id" = $1 AND i."product_id" = p."id" AND p."user_id" = $2;


