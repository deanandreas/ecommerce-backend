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

