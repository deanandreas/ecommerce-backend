-- name: InsertReview :one
INSERT INTO "reviews" ("user_id", "product_id", "rating", "comment")
    VALUES ($1, $2, $3, $4)
RETURNING
    "id";

-- name: GetProductReviews :many
SELECT
    "id",
    "user_id",
    "rating",
    "comment"
FROM
    "reviews"
WHERE
    "product_id" = $1
    AND "comment" IS NOT NULL;

-- name: GetUserReview :one
SELECT
    "id",
    "rating",
    "comment"
FROM
    "reviews"
WHERE
    "user_id" = $1
    AND "product_id" = $2;

-- name: GetReviewByID :one
SELECT
    "id",
    "rating",
    "comment"
FROM
    "reviews"
WHERE
    "id" = $1;

-- name: GetLatestReviews :many
SELECT
    r."id",
    r."rating",
    r."comment",
    r."created_at",
    r."product_id",
    p."title" AS "product_title",
    i."image_url"
FROM
    "reviews" r
    JOIN "products" p ON p."id" = r."product_id"
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
    r."comment" IS NOT NULL
ORDER BY
    r."created_at" DESC
LIMIT $1;

-- name: UpdateReview :one
UPDATE "reviews"
SET
    "rating" = $1,
    "comment" = $2,
    "updated_at" = NOW()
WHERE
    "user_id" = $3
    AND "product_id" = $4
RETURNING
    "id", "rating", "comment";

