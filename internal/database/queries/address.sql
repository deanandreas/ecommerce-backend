-- name: InsertAddress :exec
INSERT INTO "addresses" ("user_id", "street_line_1", "street_line_2", "postal_code", "state", "city", "country", "is_default")
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetAddressByUserID :many
SELECT
    *
FROM
    "addresses"
WHERE
    "user_id" = $1;

-- name: UpdateAddress :exec
UPDATE
    "addresses"
SET
    "street_line_1" = COALESCE(sqlc.narg ('street_line_1'), "street_line_1"),
    "street_line_2" = COALESCE(sqlc.narg ('street_line_2'), "street_line_2"),
    "postal_code" = COALESCE(sqlc.narg ('postal_code'), "postal_code"),
    "state" = COALESCE(sqlc.narg ('state'), "state"),
    "city" = COALESCE(sqlc.narg ('city'), "city"),
    "country" = COALESCE(sqlc.narg ('country'), "country")
WHERE
    "id" = $1
    AND "user_id" = $2;

-- name: UpdateDefaultAddress :execrows
UPDATE
    "addresses"
SET
    "is_default" = TRUE
WHERE
    "user_id" = $1
    AND "id" = $2;

-- name: DesableDefaultAddres :exec
UPDATE
    "addresses"
SET
    "is_default" = FALSE
WHERE
    "user_id" = $1;

-- name: DeleteUserAddress :execrows
DELETE FROM "addresses"
WHERE "user_id" = $1
    AND "id" = $2;

