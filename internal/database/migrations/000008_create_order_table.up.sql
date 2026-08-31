CREATE TABLE "orders" (
    "id" uuid DEFAULT gen_random_uuid (),
    "user_id" uuid NOT NULL,
    "address_id" uuid NOT NULL,
    "ship_price_in_cent" int NOT NULL,
    "status" varchar(35) DEFAULT 'pending' CHECK ("status" IN ('confirmed', 'shipped', 'cancelled', 'pending')) NOT NULL,
    "ship_date" date NOT NULL,
    "created_at" timestamp DEFAULT NOW(),
    PRIMARY KEY ("id")
);
