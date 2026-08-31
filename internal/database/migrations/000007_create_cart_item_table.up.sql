CREATE TABLE "cart_items" (
    "id" UUID DEFAULT gen_random_uuid(),
    "cart_id" UUID NOT NULL,
    "product_id" UUID NOT NULL,
    "quantity" INT NOT NULL CHECK ("quantity" > 0),
    "created_at" TIMESTAMP DEFAULT NOW(),
    "updated_at" TIMESTAMP DEFAULT NOW(),
    UNIQUE("cart_id", "product_id"),
    PRIMARY KEY ("id")
);
