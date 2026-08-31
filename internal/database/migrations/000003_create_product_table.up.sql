CREATE TABLE "products" (
    "id" UUID DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL,
    "title" TEXT NOT NULL,
    "price_in_cent" INT NOT NULL CHECK("price_in_cent" >= 0),
    "stock" INT NOT NULL CHECK("stock" >= 0),
    "category_id" UUID NOT NULL,
    "description" TEXT,
    "deleted_at" TIMESTAMP DEFAULT NULL,
    "created_at" TIMESTAMP DEFAULT NOW(),
    "updated_at" TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY ("id")
);
