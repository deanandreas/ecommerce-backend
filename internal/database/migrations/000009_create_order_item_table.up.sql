CREATE TABLE "order_items" (
    "id" UUID DEFAULT gen_random_uuid(),
    "order_id" UUID NOT NULL,
    "product_id" UUID NOT NULL,
    "quantity" INT NOT NULL,
    PRIMARY KEY ("id")
);
