-- name: GetProducts :many
SELECT
    p."id",
    p."title",
    p."price_in_cent",
    p."stock",
    i."image_url"
FROM
    "products" p
    LEFT JOIN "categories" c ON c."id" = p."category_id"
    LEFT JOIN LATERAL (
        SELECT
            "image_url"
        FROM
            "product_images"
        WHERE
            "product_id" = p."id"
            AND "is_default" = TRUE
        LIMIT 1) i ON TRUE
WHERE (sqlc.narg ('search')::text IS NULL
    OR p.title ILIKE '%' || sqlc.narg ('search')::text || '%')
AND (sqlc.narg ('slug')::text IS NULL
    OR c.slug = sqlc.narg ('slug'))
AND (sqlc.narg ('min_price')::numeric IS NULL
    OR p.price_in_cent >= sqlc.narg ('min_price'))
AND (sqlc.narg ('max_price')::numeric IS NULL
    OR p.price_in_cent <= sqlc.narg ('max_price'))
ORDER BY
    CASE WHEN sqlc.narg ('sort_by')::text = 'price_asc' THEN
        p.price_in_cent
    END ASC,
    CASE WHEN sqlc.narg ('sort_by')::text = 'price_desc' THEN
        p.price_in_cent
    END DESC,
    CASE WHEN sqlc.narg ('sort_by')::text = 'name_asc' THEN
        p.title
    END ASC,
    p.id DESC;

-- name: GetProductsCategory :many
SELECT
    *
FROM
    "categories"
LIMIT $1;

-- name: GetPopularProducts :many
SELECT
    p."id",
    p."title",
    p."price_in_cent",
    p."stock",
    i."image_url",
    ROUND(AVG(r."rating"), 2) AS "rating"
FROM
    "products" p
    LEFT JOIN "reviews" r ON r."product_id" = p."id"
    LEFT JOIN LATERAL (
        SELECT
            "image_url"
        FROM
            "product_images"
        WHERE
            "product_id" = p."id"
            AND "is_default" = TRUE
        LIMIT 1) i ON TRUE
WHERE
    "rating" > 3.5;
