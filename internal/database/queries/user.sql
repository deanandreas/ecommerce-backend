-- name: InsertUser :one
INSERT INTO "users" ("full_name", "phone", "birth", "email", "hash_password")
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    "id";

-- name: GetUserByID :one
SELECT
    u."id",
    u."full_name",
    u."phone",
    u."birth",
    u."email",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', a."id", 'street_line_1', a."street_line_1", 'street_line_2', a."street_line_2", 'postal_code', a."postal_code", 'state', a."state", 'city', a."city", 'country', a."country", 'is_default', a."is_default"))
        FROM "addresses" a
        WHERE
            a."user_id" = u."id"), '[]'::json) AS "address"
FROM
    "users" u
WHERE
    u."id" = $1
GROUP BY
    u."id";

-- name: GetUserByEmail :one
SELECT
    u."id",
    u."full_name",
    u."phone",
    u."birth",
    u."email",
    u."hash_password",
    COALESCE((
        SELECT
            JSON_AGG(JSON_BUILD_OBJECT('id', a."id", 'street_line_1', a."street_line_1", 'street_line_2', a."street_line_2", 'postal_code', a."postal_code", 'state', a."state", 'city', a."city", 'country', a."country", 'is_default', a."is_default"))
        FROM "addresses" a
        WHERE
            a."user_id" = u."id"), '[]'::json) AS "address"
FROM
    "users" u
WHERE
    u."email" = $1
GROUP BY
    u."id";

-- name: DeleteUser :exec
DELETE FROM "users"
WHERE "id" = $1;

-- name: InsertRefreshToken :exec
INSERT INTO "refresh_tokens" ("user_id", "hash_token", "expiared_at")
    VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT
    *
FROM
    "refresh_tokens"
WHERE
    "hash_token" = $1;

-- name: DeleteRefreshToken :execrows
DELETE FROM "refresh_tokens"
WHERE "id" = $1
    AND "user_id" = $2;

-- name: UpdateProfile :one
UPDATE
    "users"
SET
    "full_name" = COALESCE(sqlc.narg ('full_name'), "full_name"),
    "birth" = COALESCE(sqlc.narg ('birth'), "birth"),
    "phone" = COALESCE(sqlc.narg ('phone'), "phone"),
    "hash_password" = COALESCE(sqlc.narg ('hash_password'), "hash_password"),
    "updated_at" = NOW()
WHERE
    "id" = $1
RETURNING
    *;

