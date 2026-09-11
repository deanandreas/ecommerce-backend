-- name: InsertProduct :one
INSERT INTO "products" ("title", "user_id", "price_in_cent", "stock", "category_id", "description")
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    "id";

-- name: InsertCategory :one
INSERT INTO "categories" ("name", "slug")
    VALUES ($1, $2)
ON CONFLICT ("slug")
    DO UPDATE SET
        "name" = EXCLUDED."name"
    RETURNING
        "id";

-- name: GetProductsByUserID :many
SELECT
    p."id",
    p."title",
    p."price_in_cent",
    p."stock",
    p."description",
    COALESCE((
        SELECT
            ROW_TO_JSON(c_sub)
        FROM (
            SELECT
                c."id", c."name", c."slug"
            FROM "categories" c
            WHERE
                c."id" = p."category_id") c_sub), '{}'::json) AS "category",
    COALESCE((
        SELECT
            JSON_AGG(i."image_url")
        FROM "product_images" i
        WHERE
            i."product_id" = p."id"), '[]'::json) AS "image_url",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', r."id", 'rating', r."rating", 'comment', r."comment", 'created_at', r."created_at"))
        FROM "reviews" r
        WHERE
            r."product_id" = p."id"), '[]'::json) AS "reviews"
FROM
    "products" p
WHERE
    p."user_id" = $1
    AND "deleted_at" IS NULL;

-- name: GetProductByUserID :one
SELECT
    p."id",
    p."title",
    p."price_in_cent",
    p."stock",
    p."description",
    COALESCE((
        SELECT
            ROW_TO_JSON(c_sub)
        FROM (
            SELECT
                c."id", c."name", c."slug"
            FROM "categories" c
            WHERE
                c."id" = p."category_id") c_sub), '{}'::json) AS "category",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', i."id", 'image_url', i."image_url", 'is_default', i."is_default"))
        FROM "product_images" i
        WHERE
            i."product_id" = p."id"), '[]'::json) AS "image_url",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', r."id", 'rating', r."rating", 'comment', r."comment", 'created_at', r."created_at"))
        FROM "reviews" r
        WHERE
            r."product_id" = p."id"), '[]'::json) AS "reviews"
FROM
    "products" p
WHERE
    p."id" = $1
    AND p."user_id" = $2
    AND "deleted_at" IS NULL;

-- name: GetProductDetails :one
SELECT
    p."id",
    p."title",
    p."price_in_cent",
    p."stock",
    p."description",
    c."name" AS "category_name",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', r."id", 'user_id', r."user_id", 'rating', r."rating", 'comment', r."comment"))
        FROM "reviews" r
        WHERE
            r."product_id" = p."id"), '[]'::json) AS "review",
    COALESCE((
        SELECT
            JSON_AGG(i."image_url")
        FROM "product_images" i
        WHERE
            i."product_id" = p."id"), '[]'::json) AS "image_url"
FROM
    "products" p
    LEFT JOIN "categories" c ON c."id" = p."category_id"
WHERE
    p."id" = $1
    AND "deleted_at" IS NULL;

-- name: GetProductRating :one
SELECT
    ROUND(AVG("rating"), 2) AS "average_rating"
FROM
    "reviews" r
WHERE
    r."product_id" = $1
GROUP BY
    r."product_id";

-- name: SoftDeleteProduct :execrows
UPDATE
    "products"
SET
    "deleted_at" = NOW()
WHERE
    "id" = $1
    AND "user_id" = $2
    AND "deleted_at" IS NULL;

-- name: GetProductStock :one
SELECT
    "stock"
FROM
    "products"
WHERE
    "id" = $1;

-- name: UpdateProductStock :exec
UPDATE
    "products"
SET
    "stock" = $1
WHERE
    "id" = $2;

-- name: UpdateProduct :exec
UPDATE
    "products"
SET
    "title" = COALESCE(sqlc.narg ('title'), "title"),
    "price_in_cent" = COALESCE(sqlc.narg ('price_in_cent'), "price_in_cent"),
    "stock" = COALESCE(sqlc.narg ('stock'), "stock"),
    "description" = COALESCE(sqlc.narg ('description'), "description"),
    "category_id" = COALESCE(sqlc.narg ('category_id'), "category_id"),
    "updated_at" = NOW()
WHERE
    "id" = $1
    AND "user_id" = $2
    AND "deleted_at" IS NULL;
