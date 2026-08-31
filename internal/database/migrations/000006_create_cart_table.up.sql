CREATE TABLE "carts" (
    "id" UUID DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL UNIQUE,
    "created_at" TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY ("id")
);
