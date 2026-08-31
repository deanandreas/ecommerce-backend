CREATE TABLE "reviews" (
    "id" UUID DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL,
    "product_id" UUID NOT NULL,
    "rating" INT NOT NULL,
    "comment" TEXT NULL,
    "created_at" TIMESTAMP DEFAULT NOW(),
    "updated_at" TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY ("id")
);
